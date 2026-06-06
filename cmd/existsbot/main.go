package main

import (
	"log"

	"github.com/segfaultuwu/exists.lol/internal/api"
	"github.com/segfaultuwu/exists.lol/internal/config"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/bot"
	"github.com/segfaultuwu/exists.lol/internal/registry"
)

func main() {
	cfg := config.Load()

	// Initialize registry
	reg := registry.New()

	if err := reg.Reload(cfg.RegistryDir); err != nil {
		log.Printf("registry reload error: %v", err)
	}

	log.Printf("loaded domains: %d", len(reg.All()))
	log.Printf("registry dir=%q", cfg.RegistryDir)

	for _, err := range reg.LastErrors() {
		log.Printf("registry warning: %s", err)
	}

	// Create and start bot
	app := bot.New(cfg)

	// Start API server if enabled
	if cfg.API.Enabled {
		apiServer := api.New(api.Config{
			Host:        cfg.API.Host,
			Port:        cfg.API.Port,
			BaseDomain:  cfg.RootDomain,
			RootDomain:  cfg.RootDomain,
			RegistryDir: cfg.RegistryDir,
			APIToken:    cfg.API.Token,
			GitHubToken: cfg.GitHubToken,
			GitHubOwner: cfg.GitHubOwner,
			GitHubRepo:  cfg.GitHubRepo,
		}, reg)

		go func() {
			log.Printf("api listening on %s:%d", cfg.API.Host, cfg.API.Port)

			if err := apiServer.Start(); err != nil {
				log.Printf("api error: %v", err)
			}
		}()
	}

	// Start bot
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
