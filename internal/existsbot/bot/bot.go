package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/config"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/githubx"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/registry"
	users "github.com/segfaultuwu/exists.lol/internal/links"
)

type Bot struct {
	cfg config.Config

	gh       *githubx.Client
	dg       *discordgo.Session
	users    *users.Store
	registry *registry.Registry
}

func New(cfg config.Config) *Bot {
	userStore, err := users.Open(cfg.UsersDBPath)
	if err != nil {
		log.Fatal("failed to open users db:", err)
	}

	reg := registry.New()
	if err := reg.Reload(cfg.RegistryDir); err != nil {
		_ = userStore.Close()
		log.Fatal("failed to load registry:", err)
	}

	return &Bot{
		cfg: cfg,

		gh: githubx.New(
			cfg.GitHubToken,
			cfg.GitHubOwner,
			cfg.GitHubRepo,
		),

		users:    userStore,
		registry: reg,
	}
}

func (b *Bot) Run() error {
	session, err := discordgo.New("Bot " + b.cfg.DiscordToken)
	if err != nil {
		return err
	}

	b.dg = session

	session.AddHandler(b.onInteraction)

	if err := session.Open(); err != nil {
		return err
	}
	defer session.Close()
	defer b.users.Close()

	if err := b.registerCommands(); err != nil {
		return err
	}

	log.Println("existsbot is running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")

	return nil
}

func (b *Bot) registerCommands() error {
	for _, cmd := range Commands() {
		_, err := b.dg.ApplicationCommandCreate(
			b.cfg.DiscordAppID,
			b.cfg.DiscordGuild,
			cmd,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
