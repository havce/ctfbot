package discord

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
)

// AdminOnly restricts access to the routes to Administrators only.
var AdminOnly handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) error {
		if e.Member().Permissions.Has(discord.PermissionAdministrator) {
			return next(e)
		}

		return e.Respond(discord.InteractionResponseTypeCreateMessage,
			discord.NewMessageCreateBuilder().
				SetEphemeral(true).
				SetEmbeds(messageEmbedError("You're not authorized to run this command.")).Build())
	}
}

func (s *Server) MustBeInsideCTFAndAdmin(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) error {
		return AdminOnly(s.MustBeInsideCTF(next))(e)
	}
}

// MustBeInsideCTF is a middleware that checks whether the event
// comes from a registered CTF. Otherwise it fails.
func (s *Server) MustBeInsideCTF(next handler.Handler) handler.Handler {
	return func(e *events.InteractionCreate) (err error) {
		parent, err := s.parentChannel(e.Channel().ID())
		if err != nil {
			return err
		}

		_, err = s.CTFService.FindCTFByName(context.TODO(), parent.Name())
		if err != nil {
			_ = e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetEphemeral(true).
					SetEmbeds(messageEmbedError("You're not inside a CTF, you cannot issue this command here.")).Build())
			return err
		}

		return next(e)
	}
}
