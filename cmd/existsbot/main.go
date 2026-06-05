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

	app := bot.New(cfg)
	if cfg.API.Enabled {
		apiServer := api.New(
			cfg.API.Host,
			cfg.API.Port,
			"exists.lol",
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
