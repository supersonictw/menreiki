package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCurrentTimeTool(t *testing.T) {
	tool := NewCurrentTimeTool()
	if tool.Name() != "current_time" {
		t.Errorf("expected tool name current_time, got %s", tool.Name())
	}

	def := tool.Definition()
	if def.Function.Name != "current_time" {
		t.Errorf("expected definition name current_time, got %s", def.Function.Name)
	}

	ctx := context.Background()
	resStr, err := tool.Execute(ctx, `{"locale": "zh-TW", "timeZone": "Asia/Taipei"}`)
	if err != nil {
		t.Fatalf("unexpected error executing tool: %v", err)
	}

	var res currentTimeResult
	if err := json.Unmarshal([]byte(resStr), &res); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if res.ISO == "" {
		t.Error("expected non-empty ISO string")
	}
	if res.EpochMs == 0 {
		t.Error("expected non-zero EpochMs")
	}
	if res.TimeZone != "Asia/Taipei" {
		t.Errorf("expected timeZone Asia/Taipei, got %s", res.TimeZone)
	}
	if res.Locale != "zh-TW" {
		t.Errorf("expected locale zh-TW, got %s", res.Locale)
	}
}
