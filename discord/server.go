package discord

import (
	"context"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/snowflake/v2"
)

type Server struct {
	AppID    string
	GuildID  string
	BotToken string

	router handler.Router
	client bot.Client
}

func NewServer() *Server {
	s := &Server{
		router: handler.New(),
	}

	s.router.Use(middleware.Logger)
	s.router.Group(func(r handler.Router) {
		r.Command("/new_ctf", s.handleCommandNewCTF)
		r.Component("/new_ctf/{data}/create", s.handleCreateCTF)
	})

	s.router.Group(func(r handler.Router) {
		r.Component("/join/{ctf}", s.handleJoinCTF)
	})
	return s
}

func (s *Server) Open(ctx context.Context) (err error) {
	s.client, err = disgo.New(
		s.BotToken,
		bot.WithDefaultGateway(),
		bot.WithEventListeners(s.router),
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
