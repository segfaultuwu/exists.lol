package main

import (
	"log"

	"github.com/segfaultuwu/exists.lol/internal/existsbot/bot"
	"github.com/segfaultuwu/exists.lol/internal/existsbot/config"
)

func main() {
	cfg := config.Load()

	app := bot.New(cfg)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
