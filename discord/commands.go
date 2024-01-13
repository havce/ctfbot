package discord

import "github.com/disgoorg/disgo/discord"

var (
	commands = []discord.ApplicationCommandCreate{
		discord.SlashCommandCreate{
			Name:        "new_ctf",
			Description: "[admin] Creates a new CTF, and creates the category",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "CTF name",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "close",
			Description: "[admin] Close registrations to the named CTF",
		},
		discord.SlashCommandCreate{
			Name:        "open",
			Description: "[admin] Open registrations to the named CTF",
		},
		discord.SlashCommandCreate{
			Name:        "info",
			Description: "Information on upcoming CTFs",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "weeks",
					Description: "How many weeks away to search available CTFs.",
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "vote",
			Description: "[admin] Prompt voting on upcoming CTFs",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "weeks",
					Description: "How many weeks away to search available CTFs.",
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "flag",
			Description: "Prepends the channel name with a 🚩 emoji.",
		},
		discord.SlashCommandCreate{
			Name:        "blood",
			Description: "Prepends the channel name with a 🩸 emoji.",
		},
		discord.SlashCommandCreate{
			Name:        "new",
			Description: "Add new private text channel",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "Enable voting",
					Required:    true,
				}},
		},
	}
)
