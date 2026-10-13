package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sashabaranov/go-openai"
)

type CurrentTimeTool struct{}

func NewCurrentTimeTool() *CurrentTimeTool {
	return &CurrentTimeTool{}
}

func (t *CurrentTimeTool) Name() string {
	return "current_time"
}

func (t *CurrentTimeTool) Definition() openai.Tool {
	params := json.RawMessage(`{
		"type": "object",
		"properties": {
			"locale": {
				"type": "string",
				"description": "BCP-47 locale tag, e.g. zh-TW or en-US."
			},
			"timeZone": {
				"type": "string",
				"description": "IANA timezone, e.g. Asia/Taipei."
			}
		}
	}`)
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        t.Name(),
			Description: "Retrieves the current date and time. Use before referencing schedules, deadlines, or timestamps.",
			Parameters:  params,
		},
	}
}

type currentTimeArgs struct {
	Locale   string `json:"locale"`
	TimeZone string `json:"timeZone"`
}

type currentTimeResult struct {
	ISO       string `json:"iso"`
	EpochMs   int64  `json:"epochMs"`
	Formatted string `json:"formatted"`
	Locale    string `json:"locale"`
	TimeZone  string `json:"timeZone"`
}

func (t *CurrentTimeTool) Execute(ctx context.Context, arguments string) (string, error) {
	var args currentTimeArgs
	if arguments != "" {
		_ = json.Unmarshal([]byte(arguments), &args)
	}

	loc := time.Local
	tzName := loc.String()
	if args.TimeZone != "" {
		if l, err := time.LoadLocation(args.TimeZone); err == nil {
			loc = l
			tzName = args.TimeZone
		}
	}

	locale := args.Locale
	if locale == "" {
		locale = "zh-TW"
	}

	now := time.Now().In(loc)
	formatted := now.Format("2006-01-02 15:04:05 MST")

	res := currentTimeResult{
		ISO:       now.Format(time.RFC3339),
		EpochMs:   now.UnixMilli(),
		Formatted: formatted,
		Locale:    locale,
		TimeZone:  tzName,
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
