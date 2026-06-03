package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	domains map[string]DomainFile
}

func New() *Registry {
	return &Registry{
		domains: make(map[string]DomainFile),
	}
}

func (r *Registry) Reload(dir string) error {
	next := make(map[string]DomainFile)
	pullCmd := exec.Command("git", "pull")
	pullCmd.Run()
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}

	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".json")

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var domain DomainFile
		if err := json.Unmarshal(data, &domain); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		if err := validateDomainFile(name, domain); err != nil {
			return fmt.Errorf("validate %s: %w", path, err)
		}

		next[name] = domain
	}

	r.mu.Lock()
	r.domains = next
	r.mu.Unlock()

	return nil
}

func (r *Registry) Get(subdomain string) (DomainFile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domain, ok := r.domains[subdomain]
	return domain, ok
}

func (r *Registry) All() map[string]DomainFile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	copy := make(map[string]DomainFile, len(r.domains))
	for k, v := range r.domains {
		copy[k] = v
	}

	return copy
}

func (r *Registry) ByDiscordID(discordID string) map[string]DomainFile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]DomainFile)

	for subdomain, domain := range r.domains {
		if domain.Owner.DiscordID == discordID {
			out[subdomain] = domain
		}
	}

	return out
}

func validateDomainFile(subdomain string, domain DomainFile) error {
	if subdomain == "" {
		return fmt.Errorf("subdomain is empty")
	}

	if domain.Owner.Username == "" {
		return fmt.Errorf("owner.username is required")
	}

	if domain.Owner.GitHubUsername == "" {
		return fmt.Errorf("owner.github_username is required")
	}

	if domain.Owner.DiscordID == "" {
		return fmt.Errorf("owner.discord_id is required")
	}

	if len(domain.Records) == 0 {
		return fmt.Errorf("records are required")
	}

	for recordType, value := range domain.Records {
		if strings.TrimSpace(recordType) == "" {
			return fmt.Errorf("record type is empty")
		}

		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("record value is empty")
		}
	}

	return nil
}
