package agent

import (
	"context"
	"errors"
	"math/rand"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type GeneralAgent struct {
	client       *openai.Client
	model        string
	systemPrompt string
	mu           sync.Mutex
	histories    map[string][]openai.ChatCompletionMessage
}

func NewGeneralAgent(cfg Config) (*GeneralAgent, error) {
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

	return &GeneralAgent{
		client:       openai.NewClientWithConfig(clientCfg),
		model:        cfg.Model,
		systemPrompt: sysPrompt,
		histories:    make(map[string][]openai.ChatCompletionMessage),
	}, nil
}

func (g *GeneralAgent) ClearHistory(chatID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.histories, chatID)
}

func (g *GeneralAgent) ChatWithAI(ctx context.Context, chatID, prompt string) (string, error) {
	g.mu.Lock()
	history := g.histories[chatID]
	g.mu.Unlock()

	userMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	}

	messages := make([]openai.ChatCompletionMessage, 0, len(history)+2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: g.systemPrompt,
	})
	messages = append(messages, history...)
	messages = append(messages, userMsg)

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    g.model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no completion choices returned")
	}

	choice := resp.Choices[rand.Intn(len(resp.Choices))]
	reply := choice.Message.Content

	assistantMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: reply,
	}

	g.mu.Lock()
	history = append(history, userMsg, assistantMsg)
	if len(history) > 30 {
		history = history[len(history)-30:]
	}
	g.histories[chatID] = history
	g.mu.Unlock()

	return reply, nil
}
