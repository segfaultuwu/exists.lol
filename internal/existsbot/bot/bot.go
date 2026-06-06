package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/segfaultuwu/exists.lol/internal/config"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/apiclient"
	"github.com/segfaultuwu/exists.lol/internal/github"
	users "github.com/segfaultuwu/exists.lol/internal/links"
	"github.com/segfaultuwu/exists.lol/internal/registry"
	"github.com/segfaultuwu/exists.lol/internal/service"
	"github.com/segfaultuwu/exists.lol/internal/version"
)

// Bot represents the Discord bot application
type Bot struct {
	cfg config.Config

	ghService *github.Client
	dg        *discordgo.Session
	users     *users.Store
	registry  *registry.Registry
	api       *apiclient.Client
	domainSvc *service.DomainService

	stopPresence chan struct{}
}

// New creates a new bot instance
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

	// Create GitHub client
	ghClient := github.New(cfg.GitHubToken, cfg.GitHubOwner, cfg.GitHubRepo)

	// Create domain service
	domainSvc := service.NewDomainService(service.DomainServiceConfig{
		GitHubToken: cfg.GitHubToken,
		GitHubOwner: cfg.GitHubOwner,
		GitHubRepo:  cfg.GitHubRepo,
		BaseDomain:  cfg.RootDomain,
		RootDomain:  cfg.RootDomain,
	}, reg)

	return &Bot{
		cfg:          cfg,
		ghService:    ghClient,
		users:        userStore,
		registry:     reg,
		api:          apiClient,
		domainSvc:    domainSvc,
		stopPresence: make(chan struct{}),
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

	b.startPresenceUpdater()

	log.Println("existsbot is running. Press Ctrl+C to stop.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")

	close(b.stopPresence)

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
			return fmt.Errorf("register command %q: %w", cmd.Name, err)
		}
	}

	return nil
}

func (b *Bot) startPresenceUpdater() {
	b.updatePresence()

	ticker := time.NewTicker(60 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b.updatePresence()

			case <-b.stopPresence:
				return
			}
		}
	}()
}

func (b *Bot) updatePresence() {
	if b.dg == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domainCount := "?"

	domains, err := b.api.ListDomains(ctx)
	if err == nil {
		domainCount = fmt.Sprintf("%d", len(domains))
	} else {
		log.Println("presence: failed to fetch domains:", err)
	}

	name := fmt.Sprintf("%s domains | %s", domainCount, versionLabel())

	if len(name) > 120 {
		name = name[:120]
	}

	err = b.dg.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: "online",
		Activities: []*discordgo.Activity{
			{
				Name: name,
				Type: discordgo.ActivityTypeWatching,
			},
		},
	})
	if err != nil {
		log.Println("failed to update presence:", err)
	}
}

func versionLabel() string {
	ver := strings.TrimSpace(version.Version)
	commit := strings.TrimSpace(version.Commit)

	if ver == "" || ver == "dev" {
		ver = "dev"
	}

	if commit == "" || commit == "unknown" {
		return ver
	}

	return ver + "@" + commit
}
