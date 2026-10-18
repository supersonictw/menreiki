package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/supersonictw/menreiki/internal/agent"
)

type GlobalConfig struct {
	LogLevel         string   `toml:"log_level"`
	DatabaseFilePath string   `toml:"database_file_path"`
	ShowBanner       *bool    `toml:"show_banner"`
	BaseURL          string   `toml:"base_url"`
	Model            string   `toml:"model"`
	APIKey           string   `toml:"api_key"`
	Mode             string   `toml:"mode"`
	Prompt           string   `toml:"prompt"`
	Temperature      *float32 `toml:"temperature"`
	MaxToolSteps     int      `toml:"max_tool_steps"`
	MaxToolCalls     int      `toml:"max_tool_calls"`
	ToolTimeoutMs    int      `toml:"tool_timeout_ms"`
	TimeoutMs        int      `toml:"timeout_ms"`
	RetryCount       int      `toml:"retry_count"`
	TavilyAPIKey     string   `toml:"tavily_api_key"`
	Presence         string   `toml:"presence"`
}

type ProfileConfig struct {
	ID           string   `toml:"id"`
	Name         string   `toml:"name"`
	Token        string   `toml:"token"`
	Personality  string   `toml:"personality"`
	Mode         string   `toml:"mode"`
	Model        string   `toml:"model"`
	Prompt       string   `toml:"prompt"`
	Presence     string   `toml:"presence"`
	AppID        string   `toml:"app_id"`
	APIKey       string   `toml:"api_key"`
	BaseURL      string   `toml:"base_url"`
	TavilyAPIKey string   `toml:"tavily_api_key"`
	Temperature  *float32 `toml:"temperature"`
	Enabled      *bool    `toml:"enabled"`
}

type ConfigFile struct {
	Global   GlobalConfig    `toml:"global"`
	Profiles []ProfileConfig `toml:"profiles"`
}

type ResolvedProfile struct {
	ID           string
	Name         string
	Personality  string
	AppID        string
	Token        string
	Presence     string
	DatabasePath string
	AgentCfg     agent.Config
}

type AppConfig struct {
	Global   GlobalConfig
	Profiles map[string]*ResolvedProfile
}

// FormatPrompt interpolates {name}, {self_name}, {personality}, {self_personality}.
func FormatPrompt(prompt, name, personality string) string {
	r := strings.NewReplacer(
		"{name}", name,
		"{self_name}", name,
		"{personality}", personality,
		"{self_personality}", personality,
	)
	return r.Replace(prompt)
}

func defaultGlobal() GlobalConfig {
	showBanner := true
	defaultTemp := float32(0.7)
	return GlobalConfig{
		LogLevel:         "INFO",
		DatabaseFilePath: "menreiki.db",
		ShowBanner:       &showBanner,
		BaseURL:          "https://api.openai.com/v1",
		Model:            "gpt-4o",
		Mode:             "general",
		Prompt:           "You are a helpful assistant.",
		Temperature:      &defaultTemp,
		MaxToolSteps:     6,
		MaxToolCalls:     10,
		ToolTimeoutMs:    30000,
		TimeoutMs:        60000,
		RetryCount:       1,
	}
}

// LoadAppConfig resolves the config file path from:
// 1. specifiedPath (CLI flag/arg)
// 2. CONFIG_FILE or MENREIKI_CONFIG environment variable
// 3. Default cwd: "config.toml"
func LoadAppConfig(specifiedPath string) (*AppConfig, error) {
	configPath := strings.TrimSpace(specifiedPath)

	if configPath == "" {
		configPath = strings.TrimSpace(os.Getenv("CONFIG_FILE"))
		if configPath == "" {
			configPath = strings.TrimSpace(os.Getenv("MENREIKI_CONFIG"))
		}
	}

	if configPath == "" {
		configPath = "config.toml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file %q not found. Please create config.toml or specify path via -c/--config or CONFIG_FILE environment variable", configPath)
		}
		return nil, fmt.Errorf("failed to read config file %q: %w", configPath, err)
	}

	var cf ConfigFile
	if err := toml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config file %q: %w", configPath, err)
	}

	if len(cf.Profiles) == 0 {
		return nil, fmt.Errorf("no [[profiles]] defined in %q", configPath)
	}

	// Merge global defaults
	global := defaultGlobal()
	if cf.Global.LogLevel != "" {
		global.LogLevel = cf.Global.LogLevel
	}
	if cf.Global.DatabaseFilePath != "" {
		global.DatabaseFilePath = cf.Global.DatabaseFilePath
	}
	if cf.Global.ShowBanner != nil {
		global.ShowBanner = cf.Global.ShowBanner
	}
	if cf.Global.BaseURL != "" {
		global.BaseURL = cf.Global.BaseURL
	}
	if cf.Global.Model != "" {
		global.Model = cf.Global.Model
	}
	if cf.Global.APIKey != "" {
		global.APIKey = cf.Global.APIKey
	} else if envKey := os.Getenv("OPENAI_API_KEY"); envKey != "" {
		global.APIKey = envKey
	}
	if cf.Global.Mode != "" {
		global.Mode = cf.Global.Mode
	}
	if cf.Global.Prompt != "" {
		global.Prompt = cf.Global.Prompt
	}
	if cf.Global.Temperature != nil {
		global.Temperature = cf.Global.Temperature
	}
	if cf.Global.MaxToolSteps > 0 {
		global.MaxToolSteps = cf.Global.MaxToolSteps
	}
	if cf.Global.MaxToolCalls > 0 {
		global.MaxToolCalls = cf.Global.MaxToolCalls
	}
	if cf.Global.ToolTimeoutMs > 0 {
		global.ToolTimeoutMs = cf.Global.ToolTimeoutMs
	}
	if cf.Global.TimeoutMs > 0 {
		global.TimeoutMs = cf.Global.TimeoutMs
	}
	if cf.Global.RetryCount > 0 {
		global.RetryCount = cf.Global.RetryCount
	}
	if cf.Global.TavilyAPIKey != "" {
		global.TavilyAPIKey = cf.Global.TavilyAPIKey
	} else if envTavily := os.Getenv("TAVILY_API_KEY"); envTavily != "" {
		global.TavilyAPIKey = envTavily
	}
	if cf.Global.Presence != "" {
		global.Presence = cf.Global.Presence
	}

	resolvedProfiles := make(map[string]*ResolvedProfile)
	for i, p := range cf.Profiles {
		if p.Enabled != nil && !*p.Enabled {
			continue
		}

		if strings.TrimSpace(p.Token) == "" {
			return nil, fmt.Errorf("profile #%d is missing token", i+1)
		}

		id := p.ID
		if id == "" {
			if p.Name != "" {
				id = p.Name
			} else {
				id = fmt.Sprintf("profile-%d", i+1)
			}
		}

		name := p.Name
		if name == "" {
			name = id
		}

		personality := p.Personality
		if personality == "" {
			personality = "helpful assistant"
		}

		prompt := p.Prompt
		if prompt == "" {
			prompt = global.Prompt
		}
		systemPrompt := FormatPrompt(prompt, name, personality)

		model := p.Model
		if model == "" {
			model = global.Model
		}

		baseURL := p.BaseURL
		if baseURL == "" {
			baseURL = global.BaseURL
		}

		apiKey := p.APIKey
		if apiKey == "" {
			apiKey = global.APIKey
		}

		mode := p.Mode
		if mode == "" {
			mode = global.Mode
		}

		temperature := *global.Temperature
		if p.Temperature != nil {
			temperature = *p.Temperature
		}

		tavilyAPIKey := p.TavilyAPIKey
		if tavilyAPIKey == "" {
			tavilyAPIKey = global.TavilyAPIKey
		}

		presence := p.Presence
		if presence == "" {
			presence = global.Presence
		}

		agentCfg := agent.Config{
			Mode:           agent.Mode(mode),
			Model:          model,
			APIKey:         apiKey,
			BaseURL:        baseURL,
			SystemPrompt:   systemPrompt,
			Temperature:    temperature,
			MaxToolSteps:   global.MaxToolSteps,
			MaxToolCalls:   global.MaxToolCalls,
			ToolTimeoutMs:  global.ToolTimeoutMs,
			AgentTimeoutMs: global.TimeoutMs,
			RetryCount:     global.RetryCount,
			TavilyAPIKey:   tavilyAPIKey,
		}

		resolvedProfiles[id] = &ResolvedProfile{
			ID:           id,
			Name:         name,
			Personality:  personality,
			AppID:        p.AppID,
			Token:        p.Token,
			Presence:     presence,
			DatabasePath: global.DatabaseFilePath,
			AgentCfg:     agentCfg,
		}
	}

	if len(resolvedProfiles) == 0 {
		return nil, fmt.Errorf("all profiles in %q are disabled", configPath)
	}

	return &AppConfig{
		Global:   global,
		Profiles: resolvedProfiles,
	}, nil
}
