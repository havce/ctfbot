package discord

import "github.com/disgoorg/disgo/discord"

var commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "new",
		Description: "[admin] Creates a new CTF.",
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
		Description: "[admin] Close registrations to the CTF you're in.",
	},
	discord.SlashCommandCreate{
		Name:        "open",
		Description: "[admin] Open registrations to the CTF you're in.",
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
		Description: "Marks the challenge as solved with a 🚩 emoji.",
	},
	discord.SlashCommandCreate{
		Name:        "blood",
		Description: "Marks the challenge as solved with a 🩸 emoji.",
	},
	discord.SlashCommandCreate{
		Name:        "delete",
		Description: "[admin] Deletes the CTF.",
	},
	discord.SlashCommandCreate{
		Name:        "chal",
		Description: "Add new text channel to discuss chal.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "name",
				Description: "Enable voting",
				Required:    true,
			},
		},
	},
}
