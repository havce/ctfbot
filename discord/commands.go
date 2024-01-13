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
		discord.SlashCommandCreate{
			Name:        "close_ctf",
			Description: "Close registrations to the named CTF",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "CTF name",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "open_ctf",
			Description: "Open registrations to the named CTF",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "name",
					Description: "CTF name",
					Required:    true,
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "info_ctf",
			Description: "Information on upcoming CTFs",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionBool{
					Name:        "vote",
					Description: "Enable voting",
				},
				discord.ApplicationCommandOptionInt{
					Name:        "weeks",
					Description: "How many weeks away to search available CTFs.",
				},
			},
		},
		discord.SlashCommandCreate{
			Name:        "flag",
			Description: "Congrats! Prepend the channel name with a 🚩 emoji.",
		},
		discord.SlashCommandCreate{
			Name:        "blood",
			Description: "Prepends the channel name with a 🩸 emoji.",
		},
		discord.SlashCommandCreate{
			Name:        "new_chal",
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
