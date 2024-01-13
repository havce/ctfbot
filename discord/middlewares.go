package discord

import (
	"context"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
)

// This middleware restricts the routes to Administrators only.
var AdminOnly handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) error {
		if e.Member().Permissions.Has(discord.PermissionAdministrator) {
			return next(e)
		}

		return e.Respond(discord.InteractionResponseTypeCreateMessage,
			discord.NewMessageCreateBuilder().
				SetContent("You're not authorized to run this command.").
				SetEphemeral(true).Build())
	}
}

func (s *Server) MustBeInsideCTF(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) (err error) {
		c, ok := s.client.Caches().Channel(e.Channel().ID())
		if !ok {
			return errors.New("not in cache")
		}

		parent, ok := s.client.Caches().Channel(*c.ParentID())
		if !ok {
			return errors.New("not in cache")
		}

		_, err = s.CTFService.FindCTFByName(context.TODO(), parent.Name())
		if err != nil {
			return e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("You're not inside a CTF, you cannot issue this command.").
					SetEphemeral(true).Build())
		}

		return next(e)
	}
}
