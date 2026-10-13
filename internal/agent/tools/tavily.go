package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

type TavilySearchTool struct {
	apiKey     string
	httpClient *http.Client
}

func NewTavilySearchTool(apiKey string) *TavilySearchTool {
	return &TavilySearchTool{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (t *TavilySearchTool) Name() string {
	return "tavily_search"
}

func (t *TavilySearchTool) Definition() openai.Tool {
	params := json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "The query to search for."
			},
			"depth": {
				"type": "string",
				"enum": ["basic", "advanced"],
				"description": "Search depth (basic or advanced)."
			},
			"includeAnswer": {
				"type": "boolean",
				"description": "Whether to include an answer."
			},
			"maxResults": {
				"type": "integer",
				"description": "Maximum number of results to return (1-10)."
			}
		},
		"required": ["query"]
	}`)
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        t.Name(),
			Description: "Use Tavily web search to retrieve up-to-date information about the real world. Great for news, facts, and citations.",
			Parameters:  params,
		},
	}
}

type tavilyArgs struct {
	Query         string `json:"query"`
	Depth         string `json:"depth"`
	IncludeAnswer *bool  `json:"includeAnswer"`
	MaxResults    int    `json:"maxResults"`
}

type tavilyReq struct {
	APIKey        string `json:"api_key"`
	Query         string `json:"query"`
	SearchDepth   string `json:"search_depth,omitempty"`
	IncludeAnswer bool   `json:"include_answer"`
	MaxResults    int    `json:"max_results,omitempty"`
}

type tavilySourceItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type tavilyResp struct {
	Answer  string             `json:"answer"`
	Results []tavilySourceItem `json:"results"`
}

type tavilyToolOutput struct {
	Answer  string             `json:"answer"`
	Sources []tavilySourceItem `json:"sources"`
}

func (t *TavilySearchTool) Execute(ctx context.Context, arguments string) (string, error) {
	if t.apiKey == "" {
		return "", fmt.Errorf("missing TAVILY_API_KEY environment variable")
	}

	var args tavilyArgs
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if args.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	depth := args.Depth
	if depth == "" {
		depth = "basic"
	}
	includeAnswer := true
	if args.IncludeAnswer != nil {
		includeAnswer = *args.IncludeAnswer
	}
	maxResults := args.MaxResults
	if maxResults <= 0 || maxResults > 10 {
		maxResults = 5
	}

	reqBody := tavilyReq{
		APIKey:        t.apiKey,
		Query:         args.Query,
		SearchDepth:   depth,
		IncludeAnswer: includeAnswer,
		MaxResults:    maxResults,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("tavily search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tavily search API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tavilyData tavilyResp
	if err := json.NewDecoder(resp.Body).Decode(&tavilyData); err != nil {
		return "", fmt.Errorf("failed to decode tavily response: %w", err)
	}

	sources := tavilyData.Results
	if len(sources) > maxResults {
		sources = sources[:maxResults]
	}

	out := tavilyToolOutput{
		Answer:  tavilyData.Answer,
		Sources: sources,
	}

	outBytes, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(outBytes), nil
}
