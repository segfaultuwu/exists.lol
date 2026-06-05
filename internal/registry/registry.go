package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	domains    map[string]DomainFile
	lastErrors []string
}

func New() *Registry {
	return &Registry{
		domains: make(map[string]DomainFile),
	}
}

func (r *Registry) Token() string {
	return os.Getenv("API_TOKEN")
}

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

	if err := gitPull(); err != nil {
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

		if err := validateDomainFile(name, domain); err != nil {
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

func (r *Registry) Get(subdomain string) (DomainFile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domain, ok := r.domains[subdomain]
	return domain, ok
}

func (r *Registry) All() map[string]DomainFile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]DomainFile, len(r.domains))

	for name, domain := range r.domains {
		out[name] = domain
	}

	return out
}

func (r *Registry) LastErrors() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, len(r.lastErrors))
	copy(out, r.lastErrors)

	return out
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

func (r *Registry) Add(dir, name string, domain DomainFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return fmt.Errorf("domain name is required")
	}

	if _, exists := r.domains[name]; exists {
		return fmt.Errorf("domain already exists")
	}

	domain.Owner.Username = strings.TrimSpace(domain.Owner.Username)
	domain.Owner.GitHubUsername = strings.TrimSpace(domain.Owner.GitHubUsername)
	domain.Owner.DiscordID = strings.TrimSpace(domain.Owner.DiscordID)
	domain.Records = normalizeRecords(domain.Records)

	if err := validateDomainFile(name, domain); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, name+".json")

	data, err := json.MarshalIndent(domain, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	r.domains[name] = domain

	return nil
}
