package discord

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/havcebot"
)

const (
	flagEmoji  = "🚩"
	bloodEmoji = "🩸"
)

const (
	DefaultChannelPrivileges = discord.PermissionsAllText | discord.PermissionsAllVoice |
		discord.PermissionUseApplicationCommands | discord.PermissionAddReactions | discord.PermissionAttachFiles | discord.PermissionEmbedLinks
)

func (s *Server) handleCommandNewCTF(event *handler.CommandEvent) error {
	ctfName := s.extractCTFName(event.SlashCommandInteractionData().String("name"))

	urlEncodedCTFName := url.PathEscape(ctfName)

	// Check if CTF is already present with the same name.
	_, err := s.CTFService.FindCTFByName(context.TODO(), ctfName)
	if err == nil {
		return Error(event, havcebot.Errorf(havcebot.ECONFLICT, "A CTF with the same name has already been created."))
	}

	_, err = event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(ColorBlurple).
			SetTitle(":white_check_mark: Confirm creation").
			SetDescriptionf("Would you like to create a new CTF named `%s`?", ctfName).
			Build()).
		SetEphemeral(true).
		AddActionRow(
			discord.NewSuccessButton("Yes, create it", fmt.Sprintf("new/%s/create", urlEncodedCTFName)),
		).
		Build(),
	)
	if err != nil {
		return Error(event, err)
	}

	return err
}

func (s *Server) extractCTFName(name string) string {
	ctftimeEvent := 0

	var err error

	numberCandidate := name
	// Try to parse CTFTime URL.
	if strings.Contains(name, "ctftime.org") {
		u, err := url.Parse(name)
		if err != nil {
			s.client.Logger().Warn("Couldn't parse URL %w", err)
			return name
		}

		ep := u.EscapedPath()
		pathComponents := strings.Split(ep, "/")

		i := slices.Index(pathComponents, "event")
		if i == -1 || i+1 >= len(ep) {
			return name
		}
		numberCandidate = pathComponents[i+1]
	}

	ctftimeEvent, err = strconv.Atoi(numberCandidate)
	if err != nil {
		return name
	}

	event, err := s.CTFTimeClient.FindEventByID(context.TODO(), ctftimeEvent)
	if err != nil {
		s.client.Logger().Warn("Couldn't fetch ctftime information %w", err)
		return name
	}

	return event.Title
}

func (s *Server) handleCommandDeleteCTF(event *handler.CommandEvent) error {
	ctf, err := s.parentChannel(event.Channel().ID())
	if err != nil {
		return Error(event, err)
	}

	ctfName := ctf.Name()

	_, err = event.CreateFollowupMessage(discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(ColorFuchsia).
			SetTitle(":warning: Confirm deletion").
			SetDescriptionf("Are you sure you want to delete `%s`? There's no undo.", ctfName).
			Build()).
		SetEphemeral(true).
		AddActionRow(
			discord.NewDangerButton("Yes, delete it", "delete/really"),
		).
		Build(),
	)
	if err != nil {
		return Error(event, err)
	}

	return err
}

func (s *Server) handleDeleteCTF(event *handler.ComponentEvent) error {
	ctf, err := s.parentChannel(event.Channel().ID())
	if err != nil {
		return Error(event, err)
	}

	ctfName := ctf.Name()

	siblings := []snowflake.ID{}
	s.client.Caches().ChannelsForEach(func(channel discord.GuildChannel) {
		if channel.ParentID() == nil {
			return
		}

		if *channel.ParentID() != ctf.ID() {
			return
		}

		siblings = append(siblings, channel.ID())
	})

	// Delete all channels.
	for _, channel := range siblings {
		if err := s.client.Rest().DeleteChannel(channel); err != nil {
			return Error(event, err)
		}
	}

	// Delete parent.
	if err := s.client.Rest().DeleteChannel(ctf.ID()); err != nil {
		return Error(event, err)
	}

	// Fetch the CTF from DB to get the role ID to delete.
	ctfFromDB, err := s.CTFService.FindCTFByName(context.TODO(), ctfName)
	if err != nil {
		return Error(event, err)
	}

	roleID, err := snowflake.Parse(ctfFromDB.RoleID)
	if err != nil {
		return Error(event, err)
	}

	// Delete the role from Discord.
	if err := s.client.Rest().DeleteRole(*event.GuildID(), roleID); err != nil {
		return Error(event, err)
	}

	// Delete the CTF from db.
	if err := s.CTFService.DeleteCTF(context.TODO(), ctfName); err != nil {
		return Error(event, err)
	}

	Respond(event, "Deletion completed", fmt.Sprintf("You successfully deleted `%s`", ctfName))
	return nil
}

func (s *Server) handleCreateCTF(event *handler.ComponentEvent) error {
	ctf, err := url.PathUnescape(event.Variables["ctf"])
	if err != nil {
		return Error(event, err)
	}

	// Check again if CTF is already present with the same name.
	_, err = s.CTFService.FindCTFByName(context.TODO(), ctf)
	if err == nil {
		return Error(event, havcebot.Errorf(havcebot.ECONFLICT, "A CTF with the same name has already been created."))
	}

	// Create role with CTF name.
	role, err := s.client.Rest().CreateRole(
		*event.GuildID(),
		discord.RoleCreate{
			Name:        ctf,
			Mentionable: true,
		},
	)
	if err != nil {
		return Error(event, err)
	}

	// Create category with the name of the CTF.
	category, err := s.client.Rest().CreateGuildChannel(
		*event.GuildID(),
		discord.GuildCategoryChannelCreate{
			Name:     ctf,
			Topic:    "new ctf",
			Position: 1,
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
		return Error(event, err)
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
			Name:     s.RegistrationChannel,
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
		return Error(event, err)
	}

	// Create recruitment message in registration text channel.
	_, err = s.client.Rest().CreateMessage(regChannel.ID(), discord.NewMessageCreateBuilder().
		SetEmbeds(discord.NewEmbedBuilder().
			SetColor(ColorBlurple).
			SetDescriptionf("Press the button to join `%s`", ctf).
			Build()).
		AddActionRow(
			discord.NewPrimaryButton(fmt.Sprintf("Join %s", ctf), fmt.Sprintf("join/%s", url.PathEscape(ctf))),
		).Build())
	if err != nil {
		return Error(event, err)
	}

	// Create general channel inside category.
	_, err = s.client.Rest().CreateGuildChannel(
		*event.GuildID(),
		discord.GuildTextChannelCreate{
			Name:     s.GeneralChannel,
			ParentID: category.ID(),
			PermissionOverwrites: []discord.PermissionOverwrite{
				discord.RolePermissionOverwrite{
					RoleID: *everyoneID,
					Deny:   discord.PermissionsAll,
				},
				discord.RolePermissionOverwrite{
					RoleID: role.ID,
					Allow:  DefaultChannelPrivileges,
				},
			},
		},
	)
	if err != nil {
		return Error(event, err)
	}

	err = s.CTFService.CreateCTF(context.TODO(), &havcebot.CTF{
		Name:  ctf,
		Start: time.Now(),
		// Parse the role.ID as uint64 and then convert
		// as string.
		RoleID:  strconv.FormatUint(uint64(role.ID), 10),
		CanJoin: true,
	})
	if err != nil {
		return Error(event, err)
	}

	_, err = event.UpdateFollowupMessage(
		event.Message.ID,
		discord.NewMessageUpdateBuilder().
			SetEmbeds(discord.NewEmbedBuilder().
				SetColor(ColorGreen).
				SetDescriptionf("CTF `%s` was successfully created!", ctf).
				Build()).
			ClearContainerComponents().
			Build())
	if err != nil {
		return err
	}

	return event.DeleteInteractionResponse()
}

func (s *Server) handleJoinCTF(event *handler.ComponentEvent) error {
	ctf, err := url.PathUnescape(event.Variables["ctf"])
	if err != nil {
		return Error(event, err)
	}

	retrievedCTF, err := s.CTFService.FindCTFByName(context.TODO(), ctf)
	if err != nil {
		return Error(event, err)
	}

	roleID, err := snowflake.Parse(retrievedCTF.RoleID)
	if err != nil {
		return Error(event, err)
	}

	if !retrievedCTF.CanJoin {
		return Error(event, havcebot.Errorf(havcebot.EUNAUTHORIZED, "Registrations are closed for `%s`. Ask an admin if you want to join.", ctf))
	}

	role, found := s.client.Caches().Role(*event.GuildID(), roleID)
	if !found {
		return Error(event,
			havcebot.Errorf(havcebot.ENOTFOUND, "Couldn't find player role for `%s`. Maybe it was deleted?", ctf))
	}

	if slices.Contains(event.Member().RoleIDs, role.ID) {
		return Error(event,
			havcebot.Errorf(havcebot.ECONFLICT, "You already joined `%s`", ctf))
	}

	// Add the roleID to the roleIDs of the user.
	roleIds := append(event.Member().RoleIDs, role.ID)

	// Actually update the user.
	_, err = s.client.Rest().UpdateMember(*event.GuildID(), event.User().ID, discord.MemberUpdate{
		Roles: &roleIds,
	})
	if err != nil {
		return Error(event, err)
	}

	Respond(event, "You've been recruited.", fmt.Sprintf("You successfully joined CTF `%s`.", ctf))
	return nil
}

func (s *Server) handleUpdateCanJoin(canJoin bool) func(event *handler.CommandEvent) error {
	return func(event *handler.CommandEvent) error {
		parentChannel, err := s.parentChannel(event.Channel().ID())
		if err != nil {
			return Error(event, err)
		}

		// If you're not inside a CTF it will output a CTF not found error.
		_, err = s.CTFService.UpdateCTF(context.TODO(), parentChannel.Name(),
			havcebot.CTFUpdate{
				CanJoin: &canJoin,
			})
		if err != nil {
			return Error(event, err)
		}

		status := "opened"
		if !canJoin {
			status = "closed"
		}

		Respond(event, "Change registration status",
			fmt.Sprintf("You successfully %s registrations for `%s`.",
				status, parentChannel.Name()))
		return nil
	}
}

func (s *Server) handleFlag(blood bool) func(event *handler.CommandEvent) error {
	return func(event *handler.CommandEvent) error {
		prefix := flagEmoji
		if blood {
			prefix = bloodEmoji
		}

		if !s.flagAllowed(event.Channel().Name()) {
			return Error(event, havcebot.Errorf(
				havcebot.EINVALID, "You cannot flag here."))
		}

		// Check if someone has already flagged this.
		if utf8.RuneCountInString(event.Channel().Name()) > 0 {
			// Decode first rune. We don't care about the byte length.
			c, _ := utf8.DecodeRuneInString(event.Channel().Name())

			blocklist := []string{flagEmoji, bloodEmoji}
			// Check against blocklist.
			if slices.Contains(blocklist, string(c)) {
				return Error(event, havcebot.Errorf(havcebot.EINVALID, "Somebody has already %s this.", prefix))
			}
		}

		// Prepend the prefix emoji.
		newName := prefix + " " + event.Channel().Name()

		// Update channel name with the prefixed emoji of flag or blood.
		_, err := s.client.Rest().UpdateChannel(event.Channel().ID(), discord.GuildTextChannelUpdate{
			Name: &newName,
		})
		if err != nil {
			return Error(event, err)
		}

		// Delete response.
		if err := event.DeleteInteractionResponse(); err != nil {
			return err
		}

		// Show everyone who flagged this! Publicly post this.
		_, err = s.client.Rest().CreateMessage(event.Channel().ID(),
			discord.NewMessageCreateBuilder().
				SetEphemeral(false).
				SetEmbeds(messageEmbedSuccess(prefix+" New flag!",
					fmt.Sprintf("%s! %s has flagged `%s`.",
						cheer(), event.User().String(), event.Channel().Name()))).
				Build())
		return err
	}
}

func (s *Server) handleNewChal(event *handler.CommandEvent) error {
	chalName := event.SlashCommandInteractionData().String("name")

	// Get parent ID of the current channel.
	parentChannel, _ := s.parentChannel(event.Channel().ID())

	// Check if there's another sibling channel with the same name.
	found := false
	s.client.Caches().ChannelsForEach(func(channel discord.GuildChannel) {
		// First check if it is our sibling.
		if channel.ParentID() == nil || *channel.ParentID() != parentChannel.ID() {
			return
		}

		// Replace blood and flag indicators. We don't want to add an
		// already solved challenge.
		cleanChanName := channel.Name()

		// Append "-" to emojis, because Discord replaces spaces with dashes.
		cleanChanName = strings.NewReplacer(
			flagEmoji+"-", "",
			bloodEmoji+"-", "").Replace(cleanChanName)

		if chalName == cleanChanName {
			found = true
			return
		}
	})

	// If so, return an error.
	if found {
		return Error(event, havcebot.Errorf(
			havcebot.ECONFLICT, "Somebody has already created `%s`.", chalName))
	}

	// We already validated the existence of parentChannel in the middleware.
	// If someone has already deleted them in the meantime, well, this sucks.
	// But the error would show up in a later call.
	ctf, _ := s.CTFService.FindCTFByName(context.TODO(), parentChannel.Name())

	// Search @everyone role ID.
	var everyoneID *snowflake.ID
	s.client.Caches().RolesForEach(*event.GuildID(), func(role discord.Role) {
		if role.Name == "@everyone" {
			everyoneID = &role.ID
		}
	})

	roleID, err := snowflake.Parse(ctf.RoleID)
	if err != nil {
		return Error(event, err)
	}

	role, found := s.client.Caches().Role(*event.GuildID(), roleID)
	if !found {
		return Error(event, havcebot.Errorf(havcebot.EINTERNAL, "Couldn't find player role for `%s`. Maybe it was deleted?", ctf.Name))
	}

	// Create the channel with our custom permissions.
	// No one but the current role members should see the channel.
	channel, err := s.client.Rest().CreateGuildChannel(*event.GuildID(), discord.GuildTextChannelCreate{
		Name:     chalName,
		ParentID: parentChannel.ID(),
		PermissionOverwrites: []discord.PermissionOverwrite{
			discord.RolePermissionOverwrite{
				RoleID: *everyoneID,
				Deny:   discord.PermissionsAll,
			},
			discord.RolePermissionOverwrite{
				RoleID: role.ID,
				Allow:  DefaultChannelPrivileges,
			},
		},
	})
	if err != nil {
		return Error(event, err)
	}

	_, err = s.client.Rest().CreateMessage(channel.ID(), discord.NewMessageCreateBuilder().SetEmbeds(messageEmbedSuccess(
		"New challenge!", fmt.Sprintf("%s has created `%s`", event.User().String(), chalName))).Build())
	if err != nil {
		return Error(event, err)
	}

	Respond(event, "New channel created", fmt.Sprintf("Successfully added channel `%s`.", chalName))
	return err
}
