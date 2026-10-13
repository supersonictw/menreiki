package agent

import (
	"testing"
)

func TestNewAgentValidation(t *testing.T) {
	_, err := NewAgent(Config{
		Mode:   ModeGeneral,
		APIKey: "",
		Model:  "gpt-4o",
	})
	if err == nil {
		t.Error("expected error when APIKey is empty")
	}

	_, err = NewAgent(Config{
		Mode:   ModeThinking,
		APIKey: "test",
		Model:  "",
	})
	if err == nil {
		t.Error("expected error when Model is empty")
	}

	_, err = NewAgent(Config{
		Mode:   "invalid",
		APIKey: "test",
		Model:  "gpt-4o",
	})
	if err == nil {
		t.Error("expected error for unknown mode")
	}
}

func TestFingerprint(t *testing.T) {
	fp1 := fingerprint("current_time", `{"locale":"zh-TW"}`)
	fp2 := fingerprint("current_time", `{"locale":"zh-TW"}`)
	if fp1 != fp2 {
		t.Errorf("fingerprints should match for identical inputs: %s vs %s", fp1, fp2)
	}

	fp3 := fingerprint("current_time", `{"locale":"en-US"}`)
	if fp1 == fp3 {
		t.Errorf("fingerprints should not match for different inputs")
	}
}
