package config

import (
	"log"
	"os"
)

type Config struct {
	DiscordToken          string
	DiscordAppID          string
	DiscordGuild          string
	DiscordRequiredRoleID string

	GitHubToken string
	GitHubOwner string
	GitHubRepo  string

	LinksPath string
}

func Load() Config {
	cfg := Config{
		DiscordToken:          os.Getenv("DISCORD_TOKEN"),
		DiscordAppID:          os.Getenv("DISCORD_APP_ID"),
		DiscordGuild:          os.Getenv("DISCORD_GUILD_ID"),
		DiscordRequiredRoleID: os.Getenv("DISCORD_REQUIRED_ROLE_ID"),

		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		GitHubOwner: os.Getenv("GITHUB_OWNER"),
		GitHubRepo:  os.Getenv("GITHUB_REPO"),

		LinksPath: os.Getenv("LINKS_PATH"),
	}

	if cfg.LinksPath == "" {
		cfg.LinksPath = "data/links.json"
	}

	must("DISCORD_TOKEN", cfg.DiscordToken)
	must("DISCORD_APP_ID", cfg.DiscordAppID)
	must("DISCORD_GUILD_ID", cfg.DiscordGuild)

	must("GITHUB_TOKEN", cfg.GitHubToken)
	must("GITHUB_OWNER", cfg.GitHubOwner)
	must("GITHUB_REPO", cfg.GitHubRepo)

	return cfg
}

func must(name, value string) {
	if value == "" {
		log.Fatalf("missing %s", name)
	}
}
