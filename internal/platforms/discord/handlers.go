package discord

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *DiscordBot) registerHandlers() {
	b.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		commandName := i.ApplicationCommandData().Name
		slog.Debug("Received interaction", "command", commandName)

		switch commandName {
		case "now":
			b.handleNow(s, i)
		case "user_id":
			b.handleUserID(s, i)
		case "clear":
			b.handleClear(s, i)
		default:
			b.respondEphemeral(s, i, "無法存取該指令")
		}
	})
}

func (b *DiscordBot) getUser(i *discordgo.InteractionCreate) *discordgo.User {
	if i.Member != nil {
		return i.Member.User
	}
	return i.User
}

func (b *DiscordBot) respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *DiscordBot) handleNow(s *discordgo.Session, i *discordgo.InteractionCreate) {
	now := time.Now().Format("2006/1/2 15:04:05")
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("現在時間是：\n%s", now),
		},
	})
}

func (b *DiscordBot) handleUserID(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := b.getUser(i)
	if user == nil {
		b.respondEphemeral(s, i, "無法取得使用者資訊。")
		return
	}
	tag := user.Username
	if user.Discriminator != "" && user.Discriminator != "0" {
		tag = fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	}
	content := fmt.Sprintf("你的Discord使用者識別碼：\n%s\n`%s`", tag, user.ID)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func (b *DiscordBot) handleClear(s *discordgo.Session, i *discordgo.InteractionCreate) {
	b.agent.ClearHistory(i.ChannelID)
	b.respondEphemeral(s, i, "已清除當前頻道的對話紀錄。")
}
