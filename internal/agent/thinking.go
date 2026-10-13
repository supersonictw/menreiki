package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/supersonictw/menreiki/internal/agent/tools"
)

type ThinkingAgent struct {
	client         *openai.Client
	model          string
	temperature    float32
	systemPrompt   string
	tools          []tools.Tool
	toolMap        map[string]tools.Tool
	openAITools    []openai.Tool
	maxToolSteps   int
	maxToolCalls   int
	toolTimeout    time.Duration
	agentTimeout   time.Duration
	retryCount     int
	mu             sync.Mutex
	histories      map[string][]openai.ChatCompletionMessage
}

func NewThinkingAgent(cfg Config) (*ThinkingAgent, error) {
	if cfg.APIKey == "" || cfg.Model == "" {
		return nil, errors.New("OPENAI_API_KEY and OPENAI_CHAT_MODEL must be set")
	}

	clientCfg := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}

	sysPrompt := cfg.SystemPrompt
	if sysPrompt == "" {
		sysPrompt = "You are a helpful assistant."
	}

	temp := cfg.Temperature
	if temp == 0 {
		temp = 0.4
	}

	maxSteps := cfg.MaxToolSteps
	if maxSteps <= 0 {
		maxSteps = 6
	}

	maxCalls := cfg.MaxToolCalls
	if maxCalls <= 0 {
		maxCalls = 10
	}

	toolTimeout := 30 * time.Second
	if cfg.ToolTimeoutMs > 0 {
		toolTimeout = time.Duration(cfg.ToolTimeoutMs) * time.Millisecond
	}

	agentTimeout := 60 * time.Second
	if cfg.AgentTimeoutMs > 0 {
		agentTimeout = time.Duration(cfg.AgentTimeoutMs) * time.Millisecond
	}

	retryCount := cfg.RetryCount
	if retryCount <= 0 {
		retryCount = 1
	}

	toolList := []tools.Tool{
		tools.NewCurrentTimeTool(),
	}
	if cfg.TavilyAPIKey != "" {
		toolList = append(toolList, tools.NewTavilySearchTool(cfg.TavilyAPIKey))
	}

	toolMap := make(map[string]tools.Tool, len(toolList))
	openAITools := make([]openai.Tool, 0, len(toolList))
	for _, t := range toolList {
		toolMap[t.Name()] = t
		openAITools = append(openAITools, t.Definition())
	}

	return &ThinkingAgent{
		client:       openai.NewClientWithConfig(clientCfg),
		model:        cfg.Model,
		temperature:  temp,
		systemPrompt: sysPrompt,
		tools:        toolList,
		toolMap:      toolMap,
		openAITools:  openAITools,
		maxToolSteps: maxSteps,
		maxToolCalls: maxCalls,
		toolTimeout:  toolTimeout,
		agentTimeout: agentTimeout,
		retryCount:   retryCount,
		histories:    make(map[string][]openai.ChatCompletionMessage),
	}, nil
}

func (t *ThinkingAgent) ClearHistory(chatID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.histories, chatID)
}

func fingerprint(name, args string) string {
	var obj any
	payload := name + ":" + args
	if err := json.Unmarshal([]byte(args), &obj); err == nil {
		if canonical, err := json.Marshal(obj); err == nil {
			payload = name + ":" + string(canonical)
		}
	}
	h := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(h[:8])
}

func (t *ThinkingAgent) executeToolWithRetry(ctx context.Context, tool tools.Tool, args string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= t.retryCount; attempt++ {
		toolCtx, cancel := context.WithTimeout(ctx, t.toolTimeout)
		res, err := tool.Execute(toolCtx, args)
		cancel()
		if err == nil {
			return res, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func (t *ThinkingAgent) ChatWithAI(ctx context.Context, chatID, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, t.agentTimeout)
	defer cancel()

	t.mu.Lock()
	history := t.histories[chatID]
	t.mu.Unlock()

	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(history)+2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: t.systemPrompt,
	})
	messages = append(messages, history...)
	messages = append(messages, userMsg)

	var finalAssistantMsg openai.ChatCompletionMessage

	for step := 0; step < t.maxToolSteps; step++ {
		req := openai.ChatCompletionRequest{
			Model:       t.model,
			Messages:    messages,
			Temperature: t.temperature,
			Tools:       t.openAITools,
		}

		resp, err := t.client.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", errors.New("no completion choices returned")
		}

		choice := resp.Choices[rand.Intn(len(resp.Choices))]
		assistantMsg := choice.Message

		// If no tool calls were requested, we are done
		if len(assistantMsg.ToolCalls) == 0 {
			finalAssistantMsg = assistantMsg
			break
		}

		messages = append(messages, assistantMsg)

		// Limit and deduplicate tool calls
		rawCalls := assistantMsg.ToolCalls
		if len(rawCalls) > t.maxToolCalls {
			rawCalls = rawCalls[:t.maxToolCalls]
		}

		type callResult struct {
			callID  string
			content string
		}

		results := make([]callResult, len(rawCalls))
		fpCache := make(map[string]string)
		var fpMu sync.Mutex

		var wg sync.WaitGroup
		for idx, tc := range rawCalls {
			fp := fingerprint(tc.Function.Name, tc.Function.Arguments)
			wg.Add(1)
			go func(i int, call openai.ToolCall, key string) {
				defer wg.Done()

				fpMu.Lock()
				cached, found := fpCache[key]
				fpMu.Unlock()

				if found {
					results[i] = callResult{callID: call.ID, content: cached}
					return
				}

				tool, exists := t.toolMap[call.Function.Name]
				var output string
				if !exists {
					output = fmt.Sprintf("Tool %s is not implemented.", call.Function.Name)
				} else {
					res, err := t.executeToolWithRetry(ctx, tool, call.Function.Arguments)
					if err != nil {
						output = fmt.Sprintf("Failed to run %s: %v", call.Function.Name, err)
					} else {
						output = res
					}
				}

				fpMu.Lock()
				fpCache[key] = output
				fpMu.Unlock()

				results[i] = callResult{callID: call.ID, content: output}
			}(idx, tc, fp)
		}
		wg.Wait()

		for _, cr := range results {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: cr.callID,
				Content:    cr.content,
			})
		}
	}

	// If finished all tool steps but no final text yet, make one final call without tools
	if finalAssistantMsg.Content == "" {
		req := openai.ChatCompletionRequest{
			Model:       t.model,
			Messages:    messages,
			Temperature: t.temperature,
		}
		resp, err := t.client.CreateChatCompletion(ctx, req)
		if err == nil && len(resp.Choices) > 0 {
			finalAssistantMsg = resp.Choices[0].Message
		}
	}

	reply := finalAssistantMsg.Content
	if reply == "" {
		reply = "I'm sorry, I couldn't generate a response."
	}

	t.mu.Lock()
	history = append(history, userMsg, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: reply,
	})
	if len(history) > 30 {
		history = history[len(history)-30:]
	}
	t.histories[chatID] = history
	t.mu.Unlock()

	return reply, nil
}
