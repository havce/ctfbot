package discord

import (
	"context"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/havce/havcebot"
	"github.com/havce/havcebot/ctftime"
)

type Server struct {
	GuildID  string
	BotToken string

	router handler.Router
	client bot.Client

	CTFService    havcebot.CTFService
	CTFTimeClient *ctftime.Client
}

func NewServer() *Server {
	s := &Server{
		router: handler.New(),
	}

	// THese routes can be issued by anyone.
	s.router.Group(func(r handler.Router) {
		r.Use(AdminOnly)
		r.Command("/new_ctf", s.handleCommandNewCTF)
		r.Component("/new_ctf/{ctf}/create", s.handleCreateCTF)
		r.Command("/close_ctf", s.handleUpdateCanJoin(false))
		r.Command("/open_ctf", s.handleUpdateCanJoin(true))
	})

	// These routes are not authenticated.
	s.router.Group(func(r handler.Router) {
		r.Use(s.MustBeInsideCTF)
		r.Component("/join/{ctf}", s.handleJoinCTF)
		r.Command("/flag", s.handleFlag("🚩"))
		r.Command("/blood", s.handleFlag("🩸"))
	})

	s.router.Group(func(r handler.Router) {
		r.Command("/info_ctf", s.handleInfoCTF)
	})

	return s
}

func (s *Server) Open(ctx context.Context) (err error) {
	s.client, err = disgo.New(
		s.BotToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds|gateway.IntentGuildMembers)),
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
