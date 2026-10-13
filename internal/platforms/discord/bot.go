package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"github.com/supersonictw/menreiki/internal/agent"
	"github.com/supersonictw/menreiki/internal/database"
	"github.com/supersonictw/menreiki/internal/utils"
)

type DiscordBot struct {
	Session  *discordgo.Session
	platform *DiscordPlatform
	agent    agent.Agent
	Database *database.AppDB
}

func InitDiscord(dp *DiscordPlatform, ag agent.Agent, db *database.AppDB) (*DiscordBot, error) {
	bot := &DiscordBot{
		Session:  dp.Session,
		platform: dp,
		agent:    ag,
		Database: db,
	}

	dp.Session.AddHandler(bot.messageCreate)
	bot.registerHandlers()

	dp.Session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsDirectMessageTyping |
		discordgo.IntentMessageContent

	if err := dp.Session.Open(); err != nil {
		return nil, err
	}

	// Auto-registration of application commands
	if err := bot.checkCommandRegistration(); err != nil {
		slog.Error("Error checking command registration", "error", err)
	}

	// Setup status message
	bot.setupStatusMessage()
	go func() {
		ticker := time.NewTicker(23 * time.Hour)
		for range ticker.C {
			bot.setupStatusMessage()
		}
	}()

	return bot, nil
}

func (b *DiscordBot) Close() error {
	return b.platform.Close()
}

func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	botID := ""
	if s.State != nil && s.State.User != nil {
		botID = s.State.User.ID
	}
	if botID != "" && m.Author.ID == botID {
		return
	}

	isMentioned := false
	for _, u := range m.Mentions {
		if u.ID == botID {
			isMentioned = true
			break
		}
	}

	// In guilds, only respond if bot is mentioned. In DMs (m.GuildID == ""), respond directly.
	if m.GuildID != "" && !isMentioned {
		return
	}

	_ = s.ChannelTyping(m.ChannelID)

	content := strings.TrimSpace(m.Content)
	if m.GuildID != "" && botID != "" {
		content = strings.ReplaceAll(content, "<@!"+botID+">", "")
		content = strings.ReplaceAll(content, "<@"+botID+">", "")
		content = strings.TrimSpace(content)
	}

	if content == "" {
		_, _ = s.ChannelMessageSendReply(
			m.ChannelID,
			"Sorry, I couldn't understand your message. Please try rephrasing it.",
			m.Reference(),
		)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	responseContent, err := b.agent.ChatWithAI(ctx, m.ChannelID, content)
	if err != nil || responseContent == "" {
		if err != nil {
			slog.Error("Error generating AI response", "error", err)
		}
		_, _ = s.ChannelMessageSendReply(
			m.ChannelID,
			"Sorry, I seem to be having trouble thinking. Please try saying it in a different way.",
			m.Reference(),
		)
		return
	}

	snippets := utils.SliceContent(responseContent, 2000, "\n")
	if len(snippets) == 0 {
		return
	}

	_, _ = s.ChannelMessageSendReply(m.ChannelID, snippets[0], m.Reference())
	for _, snippet := range snippets[1:] {
		_, _ = s.ChannelMessageSend(m.ChannelID, snippet)
	}
}

func (b *DiscordBot) checkCommandRegistration() error {
	id := ""
	if b.Session.State != nil && b.Session.State.User != nil {
		id = b.Session.State.User.ID
	}

	if id == "" {
		u, err := b.Session.User("@me")
		if err != nil {
			return fmt.Errorf("failed to fetch bot user info: %w", err)
		}
		id = u.ID
	}

	currentRevision, err := b.Database.GetCommandRevision(id)
	if err != nil {
		return err
	}

	if currentRevision < CommandRevision {
		slog.Info("Registering commands", "old_revision", currentRevision, "new_revision", CommandRevision)
		_, err := b.Session.ApplicationCommandBulkOverwrite(id, "", Commands)
		if err != nil {
			return err
		}
		slog.Info("Commands registered successfully")
		return b.Database.SetCommandRevision(id, CommandRevision)
	}

	return nil
}

func (b *DiscordBot) setupStatusMessage() {
	presenceJson := viper.GetString("DISCORD_PRESENCE")
	if presenceJson != "" {
		var statusData discordgo.UpdateStatusData
		if err := json.Unmarshal([]byte(presenceJson), &statusData); err == nil {
			if err := b.Session.UpdateStatusComplex(statusData); err == nil {
				return
			}
		}
	}

	err := b.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
		IdleSince: nil,
		Activities: []*discordgo.Activity{
			{
				Name: "咲き狂え彼岸花 染まる色に嗤え",
				Type: discordgo.ActivityTypeListening,
			},
		},
		Status: string(discordgo.StatusOnline),
	})
	if err != nil {
		slog.Warn("Error updating status", "error", err)
	}
}
