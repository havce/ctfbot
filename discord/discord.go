package discord

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/havce/ctfbot"
)

type CreateFollowupMessager interface {
	CreateFollowupMessage(messageCreate discord.MessageCreate, opts ...rest.RequestOpt) (*discord.Message, error)
	Client() bot.Client
}

func Error(event CreateFollowupMessager, err error) error {
	// Extract error code and message.
	code, message := ctfbot.ErrorCode(err), ctfbot.ErrorMessage(err)

	if code == ctfbot.EINTERNAL {
		event.Client().Logger().Error("Internal server error", code, err)
	}

	// Print user message to response.
	_, err = event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(messageEmbedError(message)).Build())
	return err
}

// messageError is a utility that builds and outputs a embed.
func messageEmbedError(message string) discord.Embed {
	return discord.NewEmbedBuilder().
		SetTitle(":octagonal_sign: There was an error while handling your request.").
		SetColor(ColorRed).
		SetDescription(message).
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
	_, _ = event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(messageEmbedSuccess(title, description)).Build())
}
