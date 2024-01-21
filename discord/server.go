package discord

import (
	"context"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"

	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/ctfbot"
	"github.com/havce/ctfbot/ctftime"
)

type Server struct {
	GuildID  string
	BotToken string

	router handler.Router
	client bot.Client

	CTFService    ctfbot.CTFService
	CTFTimeClient *ctftime.Client

	// Channel default names.
	GeneralChannel      string
	RegistrationChannel string
}

func NewServer() *Server {
	s := &Server{
		router: handler.New(),
	}

	// Admin only routes. We repeat the Middleware injection because
	// order of evaluation is important. It isn't super clean, but it works.
	s.router.Group(func(r handler.Router) {
		r.Use(AdminOnly)
		r.Command("/new", s.handleCommandNewCTF)
		r.Component("/new/{ctf}/create", s.handleCreateCTF)
		r.Command("/vote", s.handleInfoCTF(true))
	})

	// Admin only routes and must be under a registered CTF.
	s.router.Group(func(r handler.Router) {
		r.Use(s.MustBeInsideCTFAndAdmin)
		r.Command("/delete", s.handleCommandDeleteCTF)
		r.Component("/delete/really", s.handleDeleteCTF)
		r.Command("/close", s.handleUpdateCanJoin(false))
		r.Command("/open", s.handleUpdateCanJoin(true))
	})

	// These routes must be hit while inside of a CTF, but don't
	// require any admin priviledge.
	s.router.Group(func(r handler.Router) {
		r.Use(s.MustBeInsideCTF)

		r.Component("/join/{ctf}", s.handleJoinCTF)
		r.Command("/flag", s.handleFlag(false))
		r.Command("/blood", s.handleFlag(true))
		r.Command("/chal", s.handleNewChal)
	})

	// These routes can be used by anyone.
	// They won't create any public message.
	s.router.Group(func(r handler.Router) {
		r.Use(middleware.Defer(discord.InteractionTypeApplicationCommand, false, true))
		r.Use(middleware.Defer(discord.InteractionTypeComponent, false, true))
		r.Command("/info", s.handleInfoCTF(false))
	})

	return s
}

func (s *Server) Open(ctx context.Context) (err error) {
	s.client, err = disgo.New(
		s.BotToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds),
		),
		bot.WithEventListeners(s.router),
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagChannels|cache.FlagMembers|cache.FlagRoles),
		),
	)
	if err != nil {
		return err
	}

	guildID, err := snowflake.Parse(s.GuildID)
	if err != nil {
		return err
	}

	if err = handler.SyncCommands(s.client, commands, []snowflake.ID{guildID}); err != nil {
		return err
	}

	return s.client.OpenGateway(ctx)
}

func (s *Server) Close(ctx context.Context) error {
	if s.client != nil {
		s.client.Close(ctx)
	}

	return nil
}
