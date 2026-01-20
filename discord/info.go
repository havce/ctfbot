package discord

import (
	"context"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/havce/ctfbot"
	"github.com/havce/ctfbot/ctftime"
)

const (
	DefaultDisplayLimit = 9
	DefaultWeeks        = 2
)

func (s *Server) handleInfoCTF(vote bool) func(event *handler.CommandEvent) error {
	return func(event *handler.CommandEvent) error {
		weeks := DefaultWeeks
		maybeWeeks, ok := event.SlashCommandInteractionData().OptInt("weeks")
		if ok && maybeWeeks > 0 {
			weeks = maybeWeeks
		}

		now := time.Now()
		finish := time.Now().Add(time.Duration(weeks) * 24 * 7 * time.Hour)

		events, err := s.CTFTimeClient.FindEvents(context.TODO(), ctftime.EventFilter{
			Start:  &now,
			Finish: &finish,
			Limit:  DefaultDisplayLimit,
		})
		if err != nil {
			return Error(event, err)
		}

		embeds := []discord.Embed{}

		for i, event := range events {
			orga := ""
			for _, team := range event.Organizers {
				orga += team.Name + " "
			}
			orga = truncate(orga, 100)
			if orga == "" {
				orga = "Unknown"
			}

			title := event.Title
			if title == "" {
				title = "Untitled Event"
			}
			title = truncate(title, 100)

			if vote {
				title = ctfbot.Itoe(i+1) + " " + title
			}

			description := truncate(event.Description, 200)

			var thumbnail *discord.EmbedResource
			if isValidURL(event.Logo) {
				thumbnail = &discord.EmbedResource{
					URL:    event.Logo,
					Width:  100,
					Height: 100,
				}
			}

			ctfTimeURL := event.CTFTimeURL
			if !isValidURL(ctfTimeURL) {
				ctfTimeURL = "N/A"
			}

			ctfLink := event.URL
			if !isValidURL(ctfLink) {
				ctfLink = "N/A"
			}

			embed := discord.Embed{
				Title:       title,
				Description: description,
				Footer: &discord.EmbedFooter{
					Text: "Informations provided here may be incorrect or out of date",
				},
				Color:     ColorNotQuiteBlack,
				Thumbnail: thumbnail,
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
						Value: ctfTimeURL,
					},
					{
						Name:  "CTF link",
						Value: ctfLink,
					},
				},
			}

			if isValidURL(event.URL) {
				embed.URL = event.URL
			}

			embeds = append(embeds, embed)
		}

		// Create a vote through a separate REST API call not to mess with
		// our complex system of mirrors and levers to handle errors.
		if vote {
			msg, err := s.client.Rest().CreateMessage(event.Channel().ID(), discord.NewMessageCreateBuilder().
				SetEmbeds(embeds...).
				SetEphemeral(false).
				Build(),
			)
			if err != nil {
				return Error(event, err)
			}

			for i := range embeds {
				err = s.client.Rest().AddReaction(event.Channel().ID(), msg.ID, ctfbot.Itoe(i+1))
				if err != nil {
					return Error(event, err)
				}
			}

			_, err = event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
				SetContent("Happy voting! :smile:").Build())
			if err != nil {
				return Error(event, err)
			}
			return nil
		}

		_, err = event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
			SetEmbeds(embeds...).
			Build(),
		)
		if err != nil {
			return Error(event, err)
		}
		return nil
	}
}
