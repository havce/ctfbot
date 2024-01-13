package discord

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/havcebot"
)

func formatTime(t *time.Time) string {
	return fmt.Sprintf("<t:%d:F>", t.Unix())
}

func (s *Server) getParentChannel(channelID snowflake.ID) (discord.GuildChannel, error) {
	currentChannel, present := s.client.Caches().Channel(channelID)
	if !present {
		return nil, havcebot.Errorf(havcebot.ENOTFOUND, "Channel not found.")
	}
	parentChannel, present := s.client.Caches().Channel(*currentChannel.ParentID())
	if !present {
		return nil, havcebot.Errorf(havcebot.ENOTFOUND, "Parent channel of %s not found.", currentChannel.Name())
	}

	return parentChannel, nil
}

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
