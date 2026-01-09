package discord

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/ctfbot"
)

func formatTime(t *time.Time) string {
	return fmt.Sprintf("<t:%d:F>", t.Unix())
}

func (s *Server) parentChannel(channelID snowflake.ID) (discord.GuildChannel, error) {
	currentChannel, present := s.client.Caches().Channel(channelID)
	if !present {
		return nil, ctfbot.Errorf(ctfbot.ENOTFOUND, "Channel not found.")
	}
	if currentChannel.ParentID() == nil {
		return nil, ctfbot.Errorf(ctfbot.ENOTFOUND, "Channel %s is not inside a category.", currentChannel.Name())
	}
	parentChannel, present := s.client.Caches().Channel(*currentChannel.ParentID())
	if !present {
		return nil, ctfbot.Errorf(ctfbot.ENOTFOUND, "Parent channel of %s not found.", currentChannel.Name())
	}

	return parentChannel, nil
}

// cheer() is a simple function that returns a random cheer phrase.
func cheer() string {
	cheers := []string{
		"Hooray",
		"Woo-hoo",
		"Cheers",
		"Yippee",
		"Yay",
		"Let's go",
		"Hip, hip, hooray",
		"Fantastic",
		"Celebrate",
		"Party time",
	}

	return cheers[rand.Intn(len(cheers))]
}
