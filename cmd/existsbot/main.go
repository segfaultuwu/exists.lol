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

	reg := registry.New()

	if err := reg.Reload(cfg.RegistryDir); err != nil {
		log.Printf("registry reload error: %v", err)
	}

	log.Printf("loaded domains: %d", len(reg.All()))
	log.Printf("registry dir=%q", cfg.RegistryDir)

	for _, err := range reg.LastErrors() {
		log.Printf("registry warning: %s", err)
	}

	app := bot.New(cfg)

	if cfg.API.Enabled {
		apiServer := api.New(
			cfg.API.Host,
			cfg.API.Port,
			cfg.RootDomain,
			cfg.RegistryDir,
			reg,
		)

		go func() {
			log.Printf("api listening on %s:%d", cfg.API.Host, cfg.API.Port)

			if err := apiServer.Start(); err != nil {
				log.Printf("api error: %v", err)
			}
		}()
	}
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
