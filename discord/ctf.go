package discord

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/havcebot"
)

func (s *Server) handleCommandNewCTF(event *handler.CommandEvent) error {
	ctfName := event.SlashCommandInteractionData().String("name")

	return event.CreateMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(ColorBlurple).
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
		*event.GuildID(),
		discord.RoleCreate{
			Name:        ctf,
			Mentionable: true,
		},
	)

	// Create category with the name of the CTF.
	category, err := s.client.Rest().CreateGuildChannel(
		*event.GuildID(),
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

	var everyoneID *snowflake.ID
	s.client.Caches().RolesForEach(*event.GuildID(), func(role discord.Role) {
		if role.Name == "@everyone" {
			everyoneID = &role.ID
		}
	})

	// Create registration channel inside category.
	regChannel, err := s.client.Rest().CreateGuildChannel(
		*event.GuildID(),
		discord.GuildTextChannelCreate{
			Name:     "registration",
			Topic:    fmt.Sprintf("%s player registration", ctf),
			ParentID: category.ID(),
			PermissionOverwrites: []discord.PermissionOverwrite{
				discord.RolePermissionOverwrite{
					RoleID: *everyoneID,
					Allow:  discord.PermissionViewChannel | discord.PermissionReadMessageHistory,
					Deny:   discord.PermissionsAll,
				},
				discord.RolePermissionOverwrite{
					RoleID: role.ID,
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
	_, err = s.client.Rest().CreateMessage(regChannel.ID(), discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(ColorBlurple).
			SetDescriptionf("Press the button to join `%s`", ctf).
			Build()).
		AddActionRow(
			discord.NewPrimaryButton(fmt.Sprintf("Join %s", ctf), fmt.Sprintf("join/%s", ctf)),
		).Build())
	if err != nil {
		return err
	}

	// Create general channel inside category.
	_, err = s.client.Rest().CreateGuildChannel(
		*event.GuildID(),
		discord.GuildTextChannelCreate{
			Name:     "general",
			ParentID: category.ID(),
			PermissionOverwrites: []discord.PermissionOverwrite{
				discord.RolePermissionOverwrite{
					RoleID: *everyoneID,
					Deny:   discord.PermissionsAll,
				},
				discord.RolePermissionOverwrite{
					RoleID: role.ID,
					Allow:  discord.PermissionsAllText | discord.PermissionUseApplicationCommands,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	err = s.CTFService.CreateCTF(context.TODO(), &havcebot.CTF{
		Name:       ctf,
		Start:      time.Now(),
		PlayerRole: ctf,
		CanJoin:    true,
	})
	if err != nil {
		return err
	}

	return event.UpdateMessage(
		discord.NewMessageUpdateBuilder().
			SetEmbeds(discord.NewEmbedBuilder().
				SetColor(ColorGreen).
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

	retrievedCTF, err := s.CTFService.FindCTFByName(context.TODO(), ctf)
	if err != nil {
		return err
	}

	if !retrievedCTF.CanJoin {
		_, err = event.UpdateInteractionResponse(
			discord.NewMessageUpdateBuilder().
				SetEmbeds(discord.NewEmbedBuilder().
					SetColor(ColorBlurple).
					SetDescriptionf("Registrations are closed for `%s`.", ctf).
					Build()).
				Build())
		return err
	}

	var ctfRole *discord.Role
	s.client.Caches().RolesForEach(*event.GuildID(), func(role discord.Role) {
		if role.Name == retrievedCTF.PlayerRole {
			ctfRole = &role
		}
	})

	if ctfRole == nil {
		return fmt.Errorf("could not find role for CTF %s", ctf)
	}

	if slices.Contains(event.Member().RoleIDs, ctfRole.ID) {
		_, err = event.UpdateInteractionResponse(
			discord.NewMessageUpdateBuilder().
				SetEmbeds(discord.NewEmbedBuilder().
					SetColor(ColorBlurple).
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
				SetColor(ColorGreen).
				SetDescriptionf("You successfully joined CTF `%s`.", ctf).
				Build()).
			Build())
	return err
}

func (s *Server) handleUpdateCanJoin(canJoin bool) func(event *handler.CommandEvent) error {
	return func(event *handler.CommandEvent) error {
		ctf := event.SlashCommandInteractionData().String("name")

		err := event.DeferCreateMessage(true)
		if err != nil {
			return err
		}

		_, err = s.CTFService.UpdateCTF(context.TODO(), ctf, havcebot.CTFUpdate{
			CanJoin: &canJoin,
		})
		if err != nil {
			return err
		}

		_, err = event.CreateFollowupMessage(
			discord.NewMessageCreateBuilder().
				SetEphemeral(true).
				SetEmbeds(discord.NewEmbedBuilder().
					SetColor(ColorGreen).
					SetDescriptionf("You successfully set registrations for CTF `%s` to %t.", ctf, canJoin).
					Build()).
				Build())
		return err
	}
}

func (s *Server) handleFlag(prefix string) func(event *handler.CommandEvent) error {
	return func(event *handler.CommandEvent) error {
		if slices.Contains(ChannelBlocklist, event.Channel().Name()) {
			return event.CreateMessage(discord.NewMessageCreateBuilder().
				SetContentf("You cannot flag in this channel!").
				SetEphemeral(true).Build())
		}

		// Check if someone has already flagged this.
		if strings.HasPrefix(event.Channel().Name(), prefix) {
			return event.CreateMessage(discord.NewMessageCreateBuilder().
				SetContentf("Someone has already flagged this!").
				SetEphemeral(true).Build())
		}

		newName := prefix + " " + event.Channel().Name()

		_, err := s.client.Rest().UpdateChannel(event.Channel().ID(), discord.GuildTextChannelUpdate{
			Name: &newName,
		})
		if err != nil {
			return err
		}

		return event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContentf("%s %s! %s has flagged %s.", prefix,
				cheer(), event.User().String(), event.Channel().Name()).
			Build())
	}
}

func (s *Server) handleNewChal(event *handler.CommandEvent) error {
	chalName := event.SlashCommandInteractionData().String("name")

	found := false
	s.client.Caches().ChannelsForEach(func(channel discord.GuildChannel) {
		if chalName == channel.Name() {
			found = true
			return
		}
	})

	if found {
		return event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContentf("Someone has already created %s!", chalName).
			SetEphemeral(true).Build())
	}

	currentChannel, _ := s.client.Caches().Channel(event.Channel().ID())
	parentChannel, _ := s.client.Caches().Channel(*currentChannel.ParentID())

	ctf, _ := s.CTFService.FindCTFByName(context.TODO(), parentChannel.Name())

	var roleID *snowflake.ID
	var everyoneID *snowflake.ID
	s.client.Caches().RolesForEach(*event.GuildID(), func(role discord.Role) {
		if role.Name == ctf.PlayerRole {
			roleID = &role.ID
		}

		if role.Name == "@everyone" {
			everyoneID = &role.ID
		}
	})

	if roleID == nil {
		return errors.New("roleID not found")
	}

	_, err := s.client.Rest().CreateGuildChannel(*event.GuildID(), discord.GuildTextChannelCreate{
		Name:     chalName,
		ParentID: parentChannel.ID(),
		PermissionOverwrites: []discord.PermissionOverwrite{
			discord.RolePermissionOverwrite{
				RoleID: *everyoneID,
				Deny:   discord.PermissionsAll,
			},
			discord.RolePermissionOverwrite{
				RoleID: *roleID,
				Allow:  discord.PermissionsAllText | discord.PermissionUseApplicationCommands,
			},
		},
	})
	if err != nil {
		return err
	}

	return event.CreateMessage(discord.NewMessageCreateBuilder().
		SetEphemeral(true).
		SetContentf("Successfully added new channel %s.", chalName).Build())
}
