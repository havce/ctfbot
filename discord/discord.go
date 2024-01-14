package discord

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/havce/havcebot"
)

type CreateFollowupMessager interface {
	CreateFollowupMessage(messageCreate discord.MessageCreate, opts ...rest.RequestOpt) (*discord.Message, error)
}

func Error(event CreateFollowupMessager, log *slog.Logger, err error) error {
	// Extract error code and message.
	code, message := havcebot.ErrorCode(err), havcebot.ErrorMessage(err)

	if code == havcebot.EINTERNAL {
		log.Error("Internal server error", code, message)
	}

	// Print user message to response.
	event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(messageEmbedError(message)).Build())
	return err
}

// messageError is a utility that builds and outputs a embed.
func messageEmbedError(message string) discord.Embed {
	return discord.NewEmbedBuilder().
		SetTitlef(":warning: There was an error on your request.").
		SetColor(ColorRed).
		SetDescriptionf(message).
		SetField(0, "Message", message, true).Build()
}

func messageEmbedSuccess(title string, description string) discord.Embed {
	return discord.NewEmbedBuilder().
		SetTitle(title).
		SetColor(ColorGreen).
		SetDescription(description).
		Build()
}

func Respond(event CreateFollowupMessager, title string, description string) {
	event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(messageEmbedSuccess(title, description)).Build())
}
