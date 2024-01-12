package discord

import "github.com/disgoorg/disgo/discord"

var (
	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "new_ctf",
			Description: "Creates a new CTF, and creates the category",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "CTF name",
					Required:    true,
				},
			},
		},
	}
)
