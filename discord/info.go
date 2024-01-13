package discord

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/havce/havcebot"
	"github.com/havce/havcebot/ctftime"
)

func formatTime(t *time.Time) string {
	return fmt.Sprintf("<t:%d:F>", t.Unix())
}

func (s *Server) handleInfoCTF(event *handler.CommandEvent) error {
	vote, ok := event.SlashCommandInteractionData().OptBool("vote")

	// In order to enable vote it needs to be both set and enabled.
	vote = vote && ok

	weeks := 2
	maybeWeeks, ok := event.SlashCommandInteractionData().OptInt("weeks")
	if ok {
		weeks = maybeWeeks
	}

	now := time.Now()

	finish := time.Now().Add(time.Duration(weeks) * 24 * 7 * time.Hour)
	events, err := s.CTFTimeClient.FindEvents(context.TODO(), ctftime.EventFilter{
		Start:  &now,
		Finish: &finish,
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
			title = fmt.Sprintf("%s %s", havcebot.Itoe(i+1), event.Title)
		}

		embed := discord.Embed{
			Title:       title,
			Description: event.Description,
			Footer: &discord.EmbedFooter{
				Text: "Informations provided here may not be correct or up to date",
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
					Value: strconv.FormatFloat(event.Weight, 'f', -1, 64),
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

	msg, err := s.client.Rest().CreateMessage(event.Channel().ID(), discord.NewMessageCreateBuilder().
		SetEmbeds(embeds...).
		Build())
	if err != nil {
		return err
	}

	event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("Listed %d upcoming CTF, from now (%s) and %s.\nHappy CTFing! :smile:", len(embeds), formatTime(&now), formatTime(&finish)).
		ClearAllowedMentions().
		SetEphemeral(true).
		Build())

	if vote {
		for i, _ := range embeds {
			err = s.client.Rest().AddReaction(event.Channel().ID(), msg.ID, havcebot.Itoe(i+1))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
