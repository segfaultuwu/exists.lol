package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/segfaultuwu/exists.lol/internal/config"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/apiclient"
	"github.com/segfaultuwu/exists.lol/internal/githubx"
	users "github.com/segfaultuwu/exists.lol/internal/links"
	"github.com/segfaultuwu/exists.lol/internal/registry"
	"github.com/segfaultuwu/exists.lol/internal/version"
)

type Bot struct {
	cfg config.Config

	gh       *githubx.Client
	dg       *discordgo.Session
	users    *users.Store
	registry *registry.Registry
	api      *apiclient.Client
}

func New(cfg config.Config) *Bot {
	userStore, err := users.Open(cfg.UsersDBPath)
	if err != nil {
		log.Fatal("failed to open users db:", err)
	}

	apiClient := apiclient.New(cfg.API.InternalURL, cfg.API.Token)

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

		api: apiClient,
	}
}

func (b *Bot) updatePresence(s *discordgo.Session) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainCount := "?"

	domains, err := b.api.ListDomains(ctx)
	if err == nil {
		domainCount = fmt.Sprintf("%d", len(domains))
	}

	err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: "online",
		Activities: []*discordgo.Activity{
			{
				Name: fmt.Sprintf("%s domains | /help | %s | %s | %s |", domainCount, version.Version, version.Commit, version.BuildDate),
				Type: discordgo.ActivityTypeWatching,
			},
		},
	})
	if err != nil {
		log.Println("failed to update presence:", err)
	}
}

func (b *Bot) startPresenceUpdater() {
	b.updatePresence(b.dg)

	ticker := time.NewTicker(60 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.updatePresence(b.dg)
			}
		}
	}()
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

	if err := b.dg.UpdateGameStatus(0, "Watching exists.lol domains"); err != nil {
		log.Println("failed to update bot status:", err)
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
