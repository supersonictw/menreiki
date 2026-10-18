package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/supersonictw/menreiki/internal/agent"
	"github.com/supersonictw/menreiki/internal/database"
	"github.com/supersonictw/menreiki/internal/utils"
)

type BotConfig struct {
	ProfileID string
	Presence  string
}

type DiscordBot struct {
	Session   *discordgo.Session
	platform  *DiscordPlatform
	agent     agent.Agent
	Database  *database.AppDB
	Presence  string
	ProfileID string
	BotID     string
	stopChan  chan struct{}
}

func InitDiscord(dp *DiscordPlatform, ag agent.Agent, db *database.AppDB) (*DiscordBot, error) {
	return InitDiscordWithConfig(dp, ag, db, BotConfig{})
}

func InitDiscordWithConfig(dp *DiscordPlatform, ag agent.Agent, db *database.AppDB, cfg BotConfig) (*DiscordBot, error) {
	bot := &DiscordBot{
		Session:   dp.Session,
		platform:  dp,
		agent:     ag,
		Database:  db,
		Presence:  cfg.Presence,
		ProfileID: cfg.ProfileID,
		stopChan:  make(chan struct{}),
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

	u, err := dp.Session.User("@me")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bot user info: %w", err)
	}
	bot.BotID = u.ID

	// Auto-registration of application commands
	if err := bot.checkCommandRegistration(); err != nil {
		slog.Error("Error checking command registration", "profile", bot.ProfileID, "error", err)
	}

	// Setup status message
	bot.setupStatusMessage()
	go func() {
		ticker := time.NewTicker(23 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bot.setupStatusMessage()
			case <-bot.stopChan:
				return
			}
		}
	}()

	return bot, nil
}

func (b *DiscordBot) Close() error {
	select {
	case <-b.stopChan:
	default:
		close(b.stopChan)
	}
	return b.platform.Close()
}

func (b *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	botID := b.BotID
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
			slog.Error("Error generating AI response", "profile", b.ProfileID, "error", err)
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
	id := b.BotID
	currentRevision, err := b.Database.GetCommandRevision(id)
	if err != nil {
		return err
	}

	if currentRevision < CommandRevision {
		slog.Info("Registering commands", "profile", b.ProfileID, "bot_id", id, "old_revision", currentRevision, "new_revision", CommandRevision)
		_, err := b.Session.ApplicationCommandBulkOverwrite(id, "", Commands)
		if err != nil {
			return err
		}
		slog.Info("Commands registered successfully", "profile", b.ProfileID, "bot_id", id)
		return b.Database.SetCommandRevision(id, CommandRevision)
	}

	return nil
}

func (b *DiscordBot) setupStatusMessage() {
	presenceStr := b.Presence

	if presenceStr != "" {
		var statusData discordgo.UpdateStatusData
		if err := json.Unmarshal([]byte(presenceStr), &statusData); err == nil {
			if err := b.Session.UpdateStatusComplex(statusData); err == nil {
				return
			}
		} else {
			if err := b.Session.UpdateStatusComplex(discordgo.UpdateStatusData{
				Activities: []*discordgo.Activity{
					{
						Name: presenceStr,
						Type: discordgo.ActivityTypeCustom,
					},
				},
				Status: string(discordgo.StatusOnline),
			}); err == nil {
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
		slog.Warn("Error updating status", "profile", b.ProfileID, "error", err)
	}
}
