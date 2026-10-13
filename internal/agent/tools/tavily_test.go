package tools

import (
	"context"
	"testing"
)

func TestTavilySearchToolMissingKey(t *testing.T) {
	tool := NewTavilySearchTool("")
	if tool.Name() != "tavily_search" {
		t.Errorf("expected tavily_search, got %s", tool.Name())
	}

	ctx := context.Background()
	_, err := tool.Execute(ctx, `{"query": "golang"}`)
	if err == nil {
		t.Error("expected error when api key is missing")
	}
}
