package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// DiscordPlatform encapsulates discordgo.Session and implements platforms.Platform[*discordgo.Session].
type DiscordPlatform struct {
	Session *discordgo.Session
}

// New creates a new DiscordPlatform.
func New(token string) (*DiscordPlatform, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	return &DiscordPlatform{Session: dg}, nil
}

// RawClient returns the underlying *discordgo.Session.
func (d *DiscordPlatform) RawClient() *discordgo.Session {
	return d.Session
}

// ChannelType returns the platform identifier string.
func (d *DiscordPlatform) ChannelType() string {
	return "discord"
}

// SendMessage sends a message to a specific channel ID or user ID.
func (d *DiscordPlatform) SendMessage(channelID, message string) error {
	_, err := d.Session.ChannelMessageSend(channelID, message)
	if err == nil {
		return nil
	}

	dmChannel, err := d.Session.UserChannelCreate(channelID)
	if err != nil {
		return fmt.Errorf("failed to send message or create DM channel: %w", err)
	}

	_, err = d.Session.ChannelMessageSend(dmChannel.ID, message)
	return err
}

// GetUserInfo returns username and display name for a given user.
func (d *DiscordPlatform) GetUserInfo(userID string) (name, displayName string, err error) {
	user, err := d.Session.User(userID)
	if err != nil {
		return "", "", err
	}
	displayName = user.GlobalName
	if displayName == "" {
		displayName = user.Username
	}
	return user.Username, displayName, nil
}

// Close closes the Discord session.
func (d *DiscordPlatform) Close() error {
	return d.Session.Close()
}

// ResolveChannelID resolves a target user ID to a DM channel ID.
func (d *DiscordPlatform) ResolveChannelID(targetID string) (string, error) {
	dmChannel, err := d.Session.UserChannelCreate(targetID)
	if err != nil {
		return targetID, err
	}
	return dmChannel.ID, nil
}
