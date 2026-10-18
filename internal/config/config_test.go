package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/supersonictw/menreiki/internal/agent"
)

func TestFormatPrompt(t *testing.T) {
	template := "Hello, I am {name}. My persona is {personality}. Also called {self_name}."
	result := FormatPrompt(template, "Menreiki", "friendly spirit")
	expected := "Hello, I am Menreiki. My persona is friendly spirit. Also called Menreiki."
	if result != expected {
		t.Errorf("FormatPrompt mismatch:\ngot:  %q\nwant: %q", result, expected)
	}
}

func TestLoadAppConfig_ValidConfig(t *testing.T) {
	tomlContent := `
[global]
log_level = "DEBUG"
database_file_path = "shared.db"
show_banner = false
base_url = "https://custom.openai.api/v1"
model = "gpt-4o"
api_key = "global-key"
mode = "general"
prompt = "Name: {name}, Personality: {personality}"

[[profiles]]
name = "BotOne"
personality = "Cool"
token = "token-bot-1"

[[profiles]]
name = "BotTwo"
personality = "Warm"
token = "token-bot-2"
mode = "thinking"
model = "gpt-4o-mini"
tavily_api_key = "tvly-key-2"
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	appCfg, err := LoadAppConfig(configFile)
	if err != nil {
		t.Fatalf("unexpected error loading app config: %v", err)
	}

	if appCfg.Global.LogLevel != "DEBUG" {
		t.Errorf("expected global log level DEBUG, got %s", appCfg.Global.LogLevel)
	}
	if appCfg.Global.DatabaseFilePath != "shared.db" {
		t.Errorf("expected shared.db, got %s", appCfg.Global.DatabaseFilePath)
	}
	if appCfg.Global.ShowBanner == nil || *appCfg.Global.ShowBanner != false {
		t.Errorf("expected ShowBanner to be false")
	}

	if len(appCfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(appCfg.Profiles))
	}

	// Verify BotOne
	b1, exists := appCfg.Profiles["BotOne"]
	if !exists {
		t.Fatalf("BotOne not found in profiles")
	}
	if b1.Token != "token-bot-1" {
		t.Errorf("expected token-bot-1, got %s", b1.Token)
	}
	if b1.Name != "BotOne" {
		t.Errorf("expected BotOne, got %s", b1.Name)
	}
	if b1.AgentCfg.Mode != agent.ModeGeneral {
		t.Errorf("expected general mode, got %s", b1.AgentCfg.Mode)
	}
	if b1.AgentCfg.Model != "gpt-4o" {
		t.Errorf("expected gpt-4o, got %s", b1.AgentCfg.Model)
	}
	if b1.AgentCfg.APIKey != "global-key" {
		t.Errorf("expected global-key, got %s", b1.AgentCfg.APIKey)
	}
	if b1.AgentCfg.BaseURL != "https://custom.openai.api/v1" {
		t.Errorf("expected https://custom.openai.api/v1, got %s", b1.AgentCfg.BaseURL)
	}
	expectedPrompt1 := "Name: BotOne, Personality: Cool"
	if b1.AgentCfg.SystemPrompt != expectedPrompt1 {
		t.Errorf("expected prompt %q, got %q", expectedPrompt1, b1.AgentCfg.SystemPrompt)
	}

	// Verify BotTwo
	b2, exists := appCfg.Profiles["BotTwo"]
	if !exists {
		t.Fatalf("BotTwo not found in profiles")
	}
	if b2.Token != "token-bot-2" {
		t.Errorf("expected token-bot-2, got %s", b2.Token)
	}
	if b2.AgentCfg.Mode != agent.ModeThinking {
		t.Errorf("expected thinking mode, got %s", b2.AgentCfg.Mode)
	}
	if b2.AgentCfg.Model != "gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", b2.AgentCfg.Model)
	}
	if b2.AgentCfg.TavilyAPIKey != "tvly-key-2" {
		t.Errorf("expected tvly-key-2, got %s", b2.AgentCfg.TavilyAPIKey)
	}
	expectedPrompt2 := "Name: BotTwo, Personality: Warm"
	if b2.AgentCfg.SystemPrompt != expectedPrompt2 {
		t.Errorf("expected prompt %q, got %q", expectedPrompt2, b2.AgentCfg.SystemPrompt)
	}
}

func TestLoadAppConfig_FromConfigFileEnv(t *testing.T) {
	tomlContent := `
[[profiles]]
name = "EnvLoadedBot"
token = "env-token"
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "custom.toml")
	if err := os.WriteFile(configFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	t.Setenv("CONFIG_FILE", configFile)

	appCfg, err := LoadAppConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := appCfg.Profiles["EnvLoadedBot"]; !ok {
		t.Errorf("expected EnvLoadedBot to be loaded via CONFIG_FILE env")
	}
}

func TestLoadAppConfig_DefaultCwd(t *testing.T) {
	tomlContent := `
[[profiles]]
name = "CwdBot"
token = "cwd-token"
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	appCfg, err := LoadAppConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := appCfg.Profiles["CwdBot"]; !ok {
		t.Errorf("expected CwdBot to be loaded from default config.toml in cwd")
	}
}

func TestLoadAppConfig_DisabledProfile(t *testing.T) {
	tomlContent := `
[[profiles]]
name = "active"
token = "active-token"

[[profiles]]
name = "inactive"
token = "inactive-token"
enabled = false
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	appCfg, err := LoadAppConfig(configFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(appCfg.Profiles) != 1 {
		t.Fatalf("expected 1 active profile, got %d", len(appCfg.Profiles))
	}
	if _, ok := appCfg.Profiles["active"]; !ok {
		t.Errorf("expected profile 'active' to be present")
	}
}

func TestLoadAppConfig_MissingToken(t *testing.T) {
	tomlContent := `
[[profiles]]
name = "NoTokenBot"
`
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configFile, []byte(tomlContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	_, err := LoadAppConfig(configFile)
	if err == nil {
		t.Errorf("expected error for missing token, got nil")
	}
}

func TestLoadAppConfig_NonExistentFile(t *testing.T) {
	_, err := LoadAppConfig("does_not_exist.toml")
	if err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}
}

func TestLoadAppConfig_ActualConfigToml(t *testing.T) {
	if _, err := os.Stat("../../config.toml"); err != nil {
		t.Skip("config.toml not present")
	}
	appCfg, err := LoadAppConfig("../../config.toml")
	if err != nil {
		t.Fatalf("failed to load root config.toml: %v", err)
	}
	if len(appCfg.Profiles) != 4 {
		t.Errorf("expected 4 profiles in config.toml, got %d", len(appCfg.Profiles))
	}
	expectedBots := []string{"kotori", "furina", "isla", "automata"}
	for _, name := range expectedBots {
		if _, ok := appCfg.Profiles[name]; !ok {
			t.Errorf("expected profile %q in config.toml", name)
		}
	}
}
