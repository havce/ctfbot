package discord

import (
	"errors"
	"fmt"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/havcebot"
	"slices"
)

func (s *Server) handleCommandNewCTF(event *handler.CommandEvent) error {
	ctfName := event.SlashCommandInteractionData().String("name")

	return event.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(havcebot.ColorBlurple).
			SetDescriptionf("Would you like to create a new CTF named `%s`?", ctfName).
			Build()).
		SetEphemeral(true).
		AddActionRow(
			discord.NewSuccessButton("Yes, create it", fmt.Sprintf("new_ctf/%s/create", ctfName)),
		).
		Build(),
	)
}

func (s *Server) handleCreateCTF(event *handler.ComponentEvent) error {
	ctf := event.Variables["ctf"]

	// Create role with CTF name.
	role, err := s.client.Rest().CreateRole(
		snowflake.MustParse(s.GuildID),
		discord.RoleCreate{
			Name:        fmt.Sprintf("%s player", ctf),
			Mentionable: true,
		},
	)

	// Create category with the name of the CTF.
	category, err := s.client.Rest().CreateGuildChannel(
		snowflake.MustParse(s.GuildID),
		discord.GuildCategoryChannelCreate{
			Name:     ctf,
			Topic:    "new ctf",
			Position: 0,
			PermissionOverwrites: []discord.PermissionOverwrite{
				discord.RolePermissionOverwrite{
					RoleID: snowflake.ID(0),
					Deny:   discord.PermissionsAll,
					Allow:  discord.PermissionViewChannel,
				},
				discord.RolePermissionOverwrite{
					RoleID: role.ID,
					Allow:  discord.PermissionsAllText | discord.PermissionsAllVoice,
				},
			},
		})
	if err != nil {
		return err
	}

	// Create registration channel inside category.
	regChannel, err := s.client.Rest().CreateGuildChannel(
		snowflake.MustParse(s.GuildID),
		discord.GuildTextChannelCreate{
			Name:     "registration",
			Topic:    fmt.Sprintf("%s player registration", ctf),
			ParentID: category.ID(),
			PermissionOverwrites: []discord.PermissionOverwrite{
				discord.RolePermissionOverwrite{
					RoleID: snowflake.ID(0),
					Allow:  discord.PermissionViewChannel | discord.PermissionReadMessageHistory,
					Deny:   discord.PermissionsAll,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	// Create recruitment message in registration text channel.
	// TODO: We need to add a database to store the currently managed CTFs.
	_, err = s.client.Rest().CreateMessage(regChannel.ID(), discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(havcebot.ColorBlurple).
			SetDescriptionf("Press the button to join `%s`", ctf).
			Build()).
		AddActionRow(
			discord.NewPrimaryButton(fmt.Sprintf("Join %s", ctf), fmt.Sprintf("join/%s", ctf)),
		).Build())
	if err != nil {
		return err
	}

	return event.UpdateMessage(
		discord.NewMessageUpdateBuilder().
			SetEmbeds(discord.NewEmbedBuilder().
				SetColor(havcebot.ColorGreen).
				SetDescriptionf("CTF `%s` was successfully created!", ctf).
				Build()).
			ClearContainerComponents().
			Build())
}

func (s *Server) handleJoinCTF(event *handler.ComponentEvent) error {
	ctf := event.Variables["ctf"]
	if ctf == "" {
		return errors.New("empty ctf name")
	}

	err := event.DeferCreateMessage(true)
	if err != nil {
		return err
	}

	roles, err := s.client.Rest().GetRoles(snowflake.MustParse(s.GuildID))
	if err != nil {
		return err
	}

	var ctfRole *discord.Role
	for _, role := range roles {
		if role.Name == fmt.Sprintf("%s player", ctf) {
			ctfRole = &role
			break
		}
	}

	if ctfRole == nil {
		return fmt.Errorf("could not find role for CTF %s", ctf)
	}

	if slices.Contains(event.Member().RoleIDs, ctfRole.ID) {
		_, err = event.UpdateInteractionResponse(
			discord.NewMessageUpdateBuilder().
				SetEmbeds(discord.NewEmbedBuilder().
					SetColor(havcebot.ColorBlurple).
					SetDescriptionf("You already joined CTF `%s`.", ctf).
					Build()).
				Build())
		return err
	}

	roleIds := append(event.Member().RoleIDs, ctfRole.ID)

	_, err = s.client.Rest().UpdateMember(*event.GuildID(), event.User().ID, discord.MemberUpdate{
		Roles: &roleIds,
	})
	if err != nil {
		return err
	}

	_, err = event.CreateFollowupMessage(
		discord.NewMessageCreateBuilder().
			SetEphemeral(true).
			SetEmbeds(discord.NewEmbedBuilder().
				SetColor(havcebot.ColorGreen).
				SetDescriptionf("You successfully joined CTF `%s`.", ctf).
				Build()).
			Build())
	return err
}
