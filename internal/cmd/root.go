package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "menreiki",
	Short: "Menreiki is an AI agent for Discord",
	Long:  `Menreiki (面霊気) is a Discord AI agent built on a multi-agent reasoning architecture with tool-calling support.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to TOML configuration file")
}
