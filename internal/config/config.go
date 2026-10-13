package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Read from .env if it exists
	if err := viper.ReadInConfig(); err != nil {
		slog.Info("No .env file found or error reading it, relying on system environment variables")
	}

	// Also read from environment variables directly
	viper.AutomaticEnv()

	// Use a replacer to allow environment variables with underscores
	// to match nested config keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values where applicable
	viper.SetDefault("SHOW_BANNER", true)
	viper.SetDefault("DATABASE_FILE_PATH", "menreiki.db")
	viper.SetDefault("OPENAI_BASE_URL", "https://api.openai.com/v1")
	viper.SetDefault("OPENAI_CHAT_MODEL", "gpt-4o")
	viper.SetDefault("AGENT_MODE_TYPE", "general")
	viper.SetDefault("AGENT_SYSTEM_PROMPT", "You are a helpful assistant.")
	viper.SetDefault("AGENT_MAX_TOOL_STEPS", 6)
	viper.SetDefault("AGENT_MAX_TOOL_CALLS", 10)
	viper.SetDefault("AGENT_TOOL_TIMEOUT_MS", 30000)
	viper.SetDefault("AGENT_TIMEOUT_MS", 60000)
	viper.SetDefault("AGENT_RETRY_COUNT", 1)
	viper.SetDefault("LOG_LEVEL", "INFO")
}
