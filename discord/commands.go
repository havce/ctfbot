package discord

import "github.com/disgoorg/disgo/discord"

var (
	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "ping",
			Description: "Replies with pong",
		},
	}
)
