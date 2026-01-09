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
	"github.com/havce/ctfbot"
	"github.com/havce/ctfbot/ctftime"
	"github.com/havce/ctfbot/discord"
	"github.com/havce/ctfbot/sqlite"
)

// Build version, injected during build.
var (
	version string
	commit  string
)

type Config struct {
	Discord struct {
		GuildID             string `toml:"guild_id"`
		BotToken            string `toml:"bot_token"`
		RegistrationChannel string `toml:"registration_channel"`
		GeneralChannel      string `toml:"general_channel"`
	} `toml:"discord"`

	DB struct {
		DSN string `toml:"dsn"`
	} `toml:"db"`
}

const (
	DefaultDSN        = "~/ctfbot.sqlite3"
	DefaultConfigPath = "~/ctfbot.toml"
)

const (
	DefaultRegistrationChannel = "registration"
	DefaultGeneralChannel      = "general"
)

// DefaultConfig returns a new instance of Config with defaults set.
func DefaultConfig() Config {
	var config Config
	config.DB.DSN = DefaultDSN
	config.Discord.RegistrationChannel = DefaultRegistrationChannel
	config.Discord.GeneralChannel = DefaultGeneralChannel
	return config
}

func main() {
	// Propagate build information to root package to share globally.
	ctfbot.Version = version
	ctfbot.Commit = commit

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	m := NewMain()

	// Parse command line flags & load configuration.
	if err := m.ParseFlagAndConfig(ctx, os.Args[1:]); errors.Is(err, flag.ErrHelp) {
		os.Exit(1)
	} else if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := m.Run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := m.Close(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Main struct {
	Config     Config
	ConfigPath string

	DB *sqlite.DB

	Discord *discord.Server
}

func NewMain() *Main {
	return &Main{
		Discord: discord.NewServer(),
		DB:      sqlite.NewDB(""),

		Config:     DefaultConfig(),
		ConfigPath: DefaultConfigPath,
	}
}

func (m *Main) Close(ctx context.Context) error {
	if m.Discord != nil {
		_ = m.Discord.Close(ctx)
	}

	if m.DB != nil {
		return m.DB.Close()
	}

	return nil
}

func (m *Main) ParseFlagAndConfig(ctx context.Context, args []string) error {
	f := flag.NewFlagSet("ctfbotd", flag.ContinueOnError)
	f.StringVar(&m.ConfigPath, "config-path", "~/.ctfbot.toml", "config file path")
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
	if err != nil {
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

// ReadConfigFile unmarshalls config from
func ReadConfigFile(filename string) (Config, error) {
	config := DefaultConfig()
	if buf, err := os.ReadFile(filename); err != nil {
		return config, err
	} else if err := toml.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

func (m *Main) Run(ctx context.Context) (err error) {
	// Expand the DSN (in case it is in the user home directory ("~")).
	// Then open the database. This will instantiate the SQLite connection
	// and execute any pending migration files.
	if m.DB.DSN, err = expandDSN(m.Config.DB.DSN); err != nil {
		return fmt.Errorf("cannot expand dsn: %w", err)
	}
	if err := m.DB.Open(); err != nil {
		return fmt.Errorf("cannot open db: %w", err)
	}

	ctfTimeClient := ctftime.NewClient()
	ctfService := sqlite.NewCTFService(m.DB)

	m.Discord.BotToken = m.Config.Discord.BotToken
	m.Discord.GuildID = m.Config.Discord.GuildID
	m.Discord.RegistrationChannel = m.Config.Discord.RegistrationChannel
	m.Discord.GeneralChannel = m.Config.Discord.GeneralChannel

	m.Discord.CTFService = ctfService
	m.Discord.CTFTimeClient = ctfTimeClient

	if err := m.Discord.Open(ctx); err != nil {
		return err
	}

	slog.Log(ctx, slog.LevelInfo, "ctfbotd started")

	return nil
}

// expandDSN expands a datasource name. Ignores in-memory databases.
func expandDSN(dsn string) (string, error) {
	if dsn == ":memory:" {
		return dsn, nil
	}
	return expand(dsn)
}
