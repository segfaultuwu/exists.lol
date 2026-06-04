package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken          string
	DiscordAppID          string
	DiscordGuild          string
	DiscordRequiredRoleID string

	GitHubToken string
	GitHubOwner string
	GitHubRepo  string

	LinksPath   string
	UsersDBPath string

	RegistryDir string

	SelfUpdateScript string
	SystemdService   string

	RedirectCNAME string
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		DiscordToken:          os.Getenv("DISCORD_TOKEN"),
		DiscordAppID:          os.Getenv("DISCORD_APP_ID"),
		DiscordGuild:          os.Getenv("DISCORD_GUILD_ID"),
		DiscordRequiredRoleID: os.Getenv("DISCORD_REQUIRED_ROLE_ID"),

		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		GitHubOwner: os.Getenv("GITHUB_OWNER"),
		GitHubRepo:  os.Getenv("GITHUB_REPO"),

		LinksPath:   os.Getenv("LINKS_PATH"),
		UsersDBPath: os.Getenv("USERS_DB_PATH"),

		RegistryDir: os.Getenv("REGISTRY_DIR"),

		SelfUpdateScript: os.Getenv("SELF_UPDATE_SCRIPT"),
		SystemdService:   os.Getenv("SYSTEMD_SERVICE"),

		RedirectCNAME: os.Getenv("REDIRECT_CNAME"),
	}

	if cfg.LinksPath == "" {
		cfg.LinksPath = "data/links.json"
	}

	if cfg.UsersDBPath == "" {
		cfg.UsersDBPath = "./data/users.db"
	}

	if cfg.RegistryDir == "" {
		cfg.RegistryDir = "domains"
	}

	if cfg.SelfUpdateScript == "" {
		cfg.SelfUpdateScript = "./scripts/self-update.sh"
	}

	if cfg.SystemdService == "" {
		cfg.SystemdService = "existsbot"
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
