package tools

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Tool represents an executable function tool for the agent.
type Tool interface {
	Name() string
	Definition() openai.Tool
	Execute(ctx context.Context, arguments string) (string, error)
}
