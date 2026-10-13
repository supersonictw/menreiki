package agent

import (
	"context"
	"fmt"
)

type Mode string

const (
	ModeGeneral  Mode = "general"
	ModeThinking Mode = "thinking"
)

type Config struct {
	Mode         Mode
	Model        string
	APIKey       string
	BaseURL      string
	SystemPrompt string
	Temperature  float32

	// Thinking mode configs
	MaxToolSteps   int
	MaxToolCalls   int
	ToolTimeoutMs  int
	AgentTimeoutMs int
	RetryCount     int

	// Tool configs
	TavilyAPIKey string
}

type Agent interface {
	ChatWithAI(ctx context.Context, chatID, prompt string) (string, error)
	ClearHistory(chatID string)
}

func NewAgent(cfg Config) (Agent, error) {
	switch cfg.Mode {
	case ModeGeneral, "":
		return NewGeneralAgent(cfg)
	case ModeThinking:
		return NewThinkingAgent(cfg)
	default:
		return nil, fmt.Errorf("unknown agent mode: %s", cfg.Mode)
	}
}
