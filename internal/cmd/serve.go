package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/supersonictw/menreiki/internal/agent"
	"github.com/supersonictw/menreiki/internal/config"
	"github.com/supersonictw/menreiki/internal/database"
	discordplatform "github.com/supersonictw/menreiki/internal/platforms/discord"
	"github.com/supersonictw/menreiki/internal/utils"
)

var serveCmd = &cobra.Command{
	Use:   "serve [config-file]",
	Short: "Start the Menreiki Discord bot daemon",
	Run: func(cmd *cobra.Command, args []string) {
		targetFile := configFile
		if len(args) > 0 && args[0] != "" {
			targetFile = args[0]
		}

		appCfg, err := config.LoadAppConfig(targetFile)
		if err != nil {
			slog.Error("Failed to load configuration", "error", err)
			os.Exit(1)
		}

		utils.InitLogger(appCfg.Global.LogLevel)

		if appCfg.Global.ShowBanner != nil && *appCfg.Global.ShowBanner {
			hostname, _ := os.Hostname()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			slog.Info("面霊気 (Menreiki)")
			slog.Info(fmt.Sprintf("已於 \"%s\"", timestamp))
			slog.Info(fmt.Sprintf("在伺服器 \"%s\" 上啟動", hostname))
			slog.Info(fmt.Sprintf("共載入 %d 個機器人設定檔", len(appCfg.Profiles)))
		}

		dbMap := make(map[string]*database.AppDB)
		var runningBots []*discordplatform.DiscordBot

		profileIDs := make([]string, 0, len(appCfg.Profiles))
		for id := range appCfg.Profiles {
			profileIDs = append(profileIDs, id)
		}
		sort.Strings(profileIDs)

		for _, id := range profileIDs {
			prof := appCfg.Profiles[id]

			appDB, ok := dbMap[prof.DatabasePath]
			if !ok {
				var err error
				appDB, err = database.InitDB(prof.DatabasePath)
				if err != nil {
					slog.Error("Failed to initialize database", "profile", id, "path", prof.DatabasePath, "error", err)
					shutdownBots(runningBots)
					os.Exit(1)
				}
				dbMap[prof.DatabasePath] = appDB
			}

			aiAgent, err := agent.NewAgent(prof.AgentCfg)
			if err != nil {
				slog.Error("Failed to initialize AI agent", "profile", id, "error", err)
				shutdownBots(runningBots)
				os.Exit(1)
			}

			discordPlatform, err := discordplatform.New(prof.Token)
			if err != nil {
				slog.Error("Failed to create Discord platform", "profile", id, "error", err)
				shutdownBots(runningBots)
				os.Exit(1)
			}

			discordBot, err := discordplatform.InitDiscordWithConfig(
				discordPlatform,
				aiAgent,
				appDB,
				discordplatform.BotConfig{
					ProfileID: id,
					Presence:  prof.Presence,
				},
			)
			if err != nil {
				slog.Error("Failed to start Discord bot", "profile", id, "error", err)
				shutdownBots(runningBots)
				os.Exit(1)
			}

			runningBots = append(runningBots, discordBot)
			slog.Info("Bot profile started successfully",
				"profile", id,
				"name", prof.Name,
				"mode", prof.AgentCfg.Mode,
				"model", prof.AgentCfg.Model,
			)
		}

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc

		slog.Info("Gracefully shutting down all bot instances...")
		shutdownBots(runningBots)
		slog.Info("All bot instances shut down successfully.")
		os.Exit(0)
	},
}

func shutdownBots(bots []*discordplatform.DiscordBot) {
	for _, bot := range bots {
		_ = bot.Close()
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
