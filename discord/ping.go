package discord

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

func (s *Server) handlePing(event *handler.CommandEvent) error {
	return event.CreateMessage(discord.MessageCreate{
		Content: "pong",
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.NewPrimaryButton("button1", "button1/testData"),
			},
		},
	})
}
