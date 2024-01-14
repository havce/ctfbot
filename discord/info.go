package discord

import (
	"context"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/havce/havcebot"
	"github.com/havce/havcebot/ctftime"
)

const DefaultDisplayLimit = 9

func (s *Server) handleInfoCTF(vote bool) func(event *handler.CommandEvent) error {
	return func(event *handler.CommandEvent) error {
		weeks := 2
		maybeWeeks, ok := event.SlashCommandInteractionData().OptInt("weeks")
		if ok {
			weeks = maybeWeeks
		}

		if err := event.DeferCreateMessage(!vote); err != nil {
			return err
		}

		now := time.Now()
		finish := time.Now().Add(time.Duration(weeks) * 24 * 7 * time.Hour)

		events, err := s.CTFTimeClient.FindEvents(context.TODO(), ctftime.EventFilter{
			Start:  &now,
			Finish: &finish,
			Limit:  DefaultDisplayLimit,
		})
		if err != nil {
			return err
		}

		embeds := []discord.Embed{}

		for i, event := range events {
			orga := ""
			for _, team := range event.Organizers {
				orga += team.Name + " "
			}

			title := event.Title
			if vote {
				title = havcebot.Itoe(i+1) + " " + title
			}

			embed := discord.Embed{
				Title:       title,
				Description: event.Description,
				Footer: &discord.EmbedFooter{
					Text: "Informations provided here may be incorrect or out of date",
				},
				Color: ColorNotQuiteBlack,
				URL:   event.URL,
				Thumbnail: &discord.EmbedResource{
					URL:    event.Logo,
					Width:  100,
					Height: 100,
				},
				Timestamp: &now,
				Fields: []discord.EmbedField{
					{
						Name:  "Organizers",
						Value: orga,
					},
					{
						Name:  "Starts",
						Value: formatTime(&event.Start),
					},
					{
						Name:  "Ends",
						Value: formatTime(&event.Finish),
					},
					{
						Name:  "Rating",
						Value: strconv.FormatFloat(event.Weight, 'f', 2, 64),
					},
					{
						Name:  "Enrolled participants",
						Value: strconv.Itoa(event.Participants),
					},
					{
						Name:  "CTFTime",
						Value: event.CTFTimeURL,
					},
					{
						Name:  "CTF link",
						Value: event.URL,
					},
				},
			}
			embeds = append(embeds, embed)
		}

		msg, err := event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
			SetEmbeds(embeds...).
			SetEphemeral(!vote).
			Build(),
		)
		if err != nil {
			return err
		}

		// If we need to vote we now react to ourselves.
		if vote {
			for i := range embeds {
				err = s.client.Rest().AddReaction(event.Channel().ID(), msg.ID, havcebot.Itoe(i+1))
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}
