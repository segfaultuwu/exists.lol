package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Registry stores domain configurations in memory
type Registry struct {
	mu         sync.RWMutex
	domains    map[string]DomainFile
	lastErrors []string
}

// New creates a new registry
func New() *Registry {
	return &Registry{
		domains: make(map[string]DomainFile),
	}
}

// Token returns the API token from environment
func (r *Registry) Token() string {
	return os.Getenv("API_TOKEN")
}

// Reload loads domain configurations from JSON files in the specified directory
func (r *Registry) Reload(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	next := make(map[string]DomainFile)
	var reloadErrors []string

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	if err := GitPull(); err != nil {
		reloadErrors = append(reloadErrors, "git pull failed: "+err.Error())
	}

	files, err := filepath.Glob(filepath.Join(absDir, "*.json"))
	if err != nil {
		return err
	}

	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".json")

		data, err := os.ReadFile(path)
		if err != nil {
			reloadErrors = append(reloadErrors, fmt.Sprintf("%s: read failed: %v", path, err))
			continue
		}

		var domain DomainFile
		if err := json.Unmarshal(data, &domain); err != nil {
			reloadErrors = append(reloadErrors, fmt.Sprintf("%s: invalid json: %v", path, err))
			continue
		}

		if err := ValidateDomainFile(name, domain); err != nil {
			reloadErrors = append(reloadErrors, fmt.Sprintf("%s: invalid config: %v", path, err))
			continue
		}

		next[name] = domain
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastErrors = reloadErrors
	r.domains = next

	return nil
}

// Get retrieves a domain by subdomain
func (r *Registry) Get(subdomain string) (DomainFile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domain, ok := r.domains[subdomain]
	return domain, ok
}

// All returns all domains in the registry
func (r *Registry) All() map[string]DomainFile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]DomainFile, len(r.domains))

	for name, domain := range r.domains {
		out[name] = domain
	}

	return out
}

// LastErrors returns any errors that occurred during the last reload
func (r *Registry) LastErrors() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, len(r.lastErrors))
	copy(out, r.lastErrors)

	return out
}

// ByDiscordID returns all domains owned by a specific Discord user
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

// Contains checks if a subdomain exists in the registry
func (r *Registry) Contains(subdomain string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.domains[subdomain]
	return exists
}

// Count returns the number of domains in the registry
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.domains)
}
