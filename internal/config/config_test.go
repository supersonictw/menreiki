package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConfigDefaults(t *testing.T) {
	InitConfig()

	if viper.GetString("LOG_LEVEL") == "" {
		t.Errorf("expected non-empty LOG_LEVEL default")
	}
	if viper.GetString("DATABASE_FILE_PATH") == "" {
		t.Errorf("expected non-empty DATABASE_FILE_PATH default")
	}
	if viper.GetString("OPENAI_CHAT_MODEL") == "" {
		t.Errorf("expected non-empty OPENAI_CHAT_MODEL default")
	}
	if viper.GetString("AGENT_MODE_TYPE") == "" {
		t.Errorf("expected non-empty AGENT_MODE_TYPE default")
	}
}
