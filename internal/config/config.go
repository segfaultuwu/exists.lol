package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config contains all application configuration
type Config struct {
	// Discord configuration
	DiscordToken          string
	DiscordAppID          string
	DiscordGuild          string
	DiscordRequiredRoleID string

	// GitHub configuration
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string

	// Storage configuration
	LinksPath   string
	UsersDBPath string

	// Domain registry configuration
	RegistryDir string

	// System configuration
	SelfUpdateScript string
	SystemdService   string
	RedirectCNAME    string
	RootDomain       string

	// API configuration
	API APIConfig
}

// APIConfig contains API-specific configuration
type APIConfig struct {
	Enabled     bool
	Host        string
	Port        int
	PublicURL   string
	Token       string
	InternalURL string
}

// Load loads configuration from environment variables
func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		// Discord configuration
		DiscordToken:          os.Getenv("DISCORD_TOKEN"),
		DiscordAppID:          os.Getenv("DISCORD_APP_ID"),
		DiscordGuild:          os.Getenv("DISCORD_GUILD_ID"),
		DiscordRequiredRoleID: os.Getenv("DISCORD_REQUIRED_ROLE_ID"),

		// GitHub configuration
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		GitHubOwner: os.Getenv("GITHUB_OWNER"),
		GitHubRepo:  os.Getenv("GITHUB_REPO"),

		// Storage configuration
		LinksPath:   os.Getenv("LINKS_PATH"),
		UsersDBPath: os.Getenv("USERS_DB_PATH"),

		// Domain registry configuration
		RegistryDir: os.Getenv("REGISTRY_DIR"),

		// System configuration
		SelfUpdateScript: os.Getenv("SELF_UPDATE_SCRIPT"),
		SystemdService:   os.Getenv("SYSTEMD_SERVICE"),
		RedirectCNAME:    os.Getenv("REDIRECT_CNAME"),
		RootDomain:       os.Getenv("ROOT_DOMAIN"),

		// API configuration
		API: APIConfig{
			Enabled:     os.Getenv("API_ENABLED") == "true",
			Host:        envString("API_HOST", "0.0.0.0"),
			Port:        envInt("API_PORT", 8080),
			PublicURL:   envString("PUBLIC_BASE_URL", "http://localhost:8080"),
			Token:       envString("API_TOKEN", ""),
			InternalURL: envString("API_INTERNAL_URL", "http://localhost:8080"),
		},
	}

	// Set defaults
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

	// Validate required configuration
	must("DISCORD_TOKEN", cfg.DiscordToken)
	must("DISCORD_APP_ID", cfg.DiscordAppID)
	must("DISCORD_GUILD_ID", cfg.DiscordGuild)

	must("GITHUB_TOKEN", cfg.GitHubToken)
	must("GITHUB_OWNER", cfg.GitHubOwner)
	must("GITHUB_REPO", cfg.GitHubRepo)

	return cfg
}

// must ensures a configuration value is not empty
func must(name, value string) {
	if value == "" {
		log.Fatalf("missing %s", name)
	}
}

// envBool parses a boolean environment variable with fallback
func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

// envString returns an environment variable or fallback
func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// envInt parses an integer environment variable with fallback
func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
