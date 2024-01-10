package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/havce/havcebot"
	"github.com/havce/havcebot/discord"
)

// Build version, injected during build.
var (
	version string
	commit  string
)

type Config struct {
	GuildID  string `toml:"guild_id"`
	BotToken string `toml:"bot_token"`
	AppID    string `toml:"app_id"`
}

func main() {
	// Propagate build information to root package to share globally.
	havcebot.Version = strings.TrimPrefix(version, "")
	havcebot.Commit = commit

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	m := NewMain()

	// Parse command line flags & load configuration.
	if err := m.ParseFlagAndConfig(ctx, os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := m.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := m.Close(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	Config     Config
	ConfigPath string

	Discord *discord.Server
}

func NewMain() *Main {
	return &Main{
		Discord: discord.NewServer(),
	}
}

func (m *Main) Close(ctx context.Context) error {
	if m.Discord != nil {
		return m.Discord.Close(ctx)
	}

	return nil
}

func (m *Main) ParseFlagAndConfig(ctx context.Context, args []string) error {
	f := flag.NewFlagSet("havcebotd", flag.ContinueOnError)
	f.StringVar(&m.ConfigPath, "config-path", "~/.havcebot.toml", "config file path")
	if err := f.Parse(args); err != nil {
		return err
	}

	// The expand() function is here to automatically expand "~" to the user's
	// home directory. This is a common task as configuration files are typing
	// under the home directory during local development.
	configPath, err := expand(m.ConfigPath)
	if err != nil {
		return err
	}

	// Read our TOML formatted configuration file.
	config, err := ReadConfigFile(configPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", m.ConfigPath)
	} else if err != nil {
		return err
	}

	m.Config = config

	return nil
}

// expand returns path using tilde expansion. This means that a file path that
// begins with the "~" will be expanded to prefix the user's home directory.
func expand(path string) (string, error) {
	// Ignore path if it hasn't a leading tilde.
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return filepath.Clean(path), nil
	}

	// Fetch the current user to determine the home path.
	u, err := user.Current()
	if err == nil {
		return filepath.Clean(path), err
	} else if u.HomeDir == "" {
		return filepath.Clean(path), errors.New("home directory unset")
	}

	// If the path is composed only by the tilde return the home directory.
	if path == "~" {
		return u.HomeDir, nil
	}

	return filepath.Join(u.HomeDir, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
}

// ReadConfigFile unmarshals config from
func ReadConfigFile(filename string) (Config, error) {
	config := Config{}
	if buf, err := os.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

func (m *Main) Run(ctx context.Context) error {
	m.Discord.AppID = m.Config.AppID
	m.Discord.BotToken = m.Config.BotToken
	m.Discord.GuildID = m.Config.GuildID

	if err := m.Discord.Open(ctx); err != nil {
		return err
	}

	slog.Log(ctx, slog.LevelInfo, "havcebotd started")

	return nil
}
