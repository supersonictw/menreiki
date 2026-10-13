package discord

import "github.com/bwmarrin/discordgo"

const CommandRevision = 1

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "now",
		Description: "取得現在時間",
	},
	{
		Name:        "user_id",
		Description: "取得使用者ID",
	},
	{
		Name:        "clear",
		Description: "清除當前頻道的對話紀錄",
	},
}
