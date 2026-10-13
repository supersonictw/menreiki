package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/supersonictw/menreiki/internal/agent"
	"github.com/supersonictw/menreiki/internal/database"
	discordplatform "github.com/supersonictw/menreiki/internal/platforms/discord"
	"github.com/supersonictw/menreiki/internal/utils"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Menreiki Discord bot daemon",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Init DB
		databaseFilePath := viper.GetString("DATABASE_FILE_PATH")
		if databaseFilePath == "" {
			databaseFilePath = "menreiki.db"
		}

		utils.InitLogger(viper.GetString("LOG_LEVEL"))

		appDB, err := database.InitDB(databaseFilePath)
		if err != nil {
			slog.Error("Failed to initialize database", "error", err)
			os.Exit(1)
		}

		// 2. Init Agent
		agentCfg := agent.Config{
			Mode:           agent.Mode(viper.GetString("AGENT_MODE_TYPE")),
			Model:          viper.GetString("OPENAI_CHAT_MODEL"),
			APIKey:         viper.GetString("OPENAI_API_KEY"),
			BaseURL:        viper.GetString("OPENAI_BASE_URL"),
			SystemPrompt:   viper.GetString("AGENT_SYSTEM_PROMPT"),
			Temperature:    float32(viper.GetFloat64("AGENT_TEMPERATURE")),
			MaxToolSteps:   viper.GetInt("AGENT_MAX_TOOL_STEPS"),
			MaxToolCalls:   viper.GetInt("AGENT_MAX_TOOL_CALLS"),
			ToolTimeoutMs:  viper.GetInt("AGENT_TOOL_TIMEOUT_MS"),
			AgentTimeoutMs: viper.GetInt("AGENT_TIMEOUT_MS"),
			RetryCount:     viper.GetInt("AGENT_RETRY_COUNT"),
			TavilyAPIKey:   viper.GetString("TAVILY_API_KEY"),
		}

		aiAgent, err := agent.NewAgent(agentCfg)
		if err != nil {
			slog.Error("Failed to initialize AI agent", "error", err)
			os.Exit(1)
		}

		// 3. Init Discord Platform
		token := viper.GetString("DISCORD_BOT_TOKEN")
		if token == "" {
			token = viper.GetString("DISCORD_TOKEN")
		}
		if token == "" {
			slog.Error("DISCORD_BOT_TOKEN is not set in environment variables")
			os.Exit(1)
		}

		discordPlatform, err := discordplatform.New(token)
		if err != nil {
			slog.Error("Failed to create Discord platform", "error", err)
			os.Exit(1)
		}

		// 4. Init Discord Bot
		discordBot, err := discordplatform.InitDiscord(discordPlatform, aiAgent, appDB)
		if err != nil {
			slog.Error("Failed to start Discord bot", "error", err)
			os.Exit(1)
		}

		if isShowBanner := viper.GetBool("SHOW_BANNER"); isShowBanner {
			hostname, _ := os.Hostname()
			timestamp := time.Now().String()
			slog.Info("面霊気 (Menreiki)")
			slog.Info(fmt.Sprintf("已於 \"%s\"", timestamp))
			slog.Info(fmt.Sprintf("在伺服器 \"%s\" 上啟動", hostname))
		}

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc

		_ = discordBot.Close()
		slog.Info("Gracefully shutting down.")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
