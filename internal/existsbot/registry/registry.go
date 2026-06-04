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
	mu         sync.RWMutex
	domains    map[string]DomainFile
	lastErrors []string
}

func New() *Registry {
	return &Registry{
		domains: make(map[string]DomainFile),
	}
}

func (r *Registry) Reload(dir string) error {
	next := make(map[string]DomainFile)
	var reloadErrors []string

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	if err := gitPull(); err != nil {
		reloadErrors = append(reloadErrors, "git pull failed: "+err.Error())
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
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

	if len(files) > 0 && len(next) == 0 {
		return nil
	}

	r.domains = next

	return nil
}

func gitPull() error {
	cmd := exec.Command("git", "pull", "--ff-only")

	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return err
		}

		return fmt.Errorf("%w: %s", err, msg)
	}

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

func validateDomainFile(subdomain string, domain DomainFile) error {
	if subdomain == "" {
		return fmt.Errorf("subdomain is empty")
	}

	if strings.TrimSpace(domain.Owner.Username) == "" {
		return fmt.Errorf("owner.username is required")
	}

	if strings.TrimSpace(domain.Owner.GitHubUsername) == "" {
		return fmt.Errorf("owner.github_username is required")
	}

	if strings.TrimSpace(domain.Owner.DiscordID) == "" {
		return fmt.Errorf("owner.discord_id is required")
	}

	if len(domain.Records) == 0 {
		return fmt.Errorf("records are required")
	}

	for recordType, values := range domain.Records {
		recordType = strings.ToUpper(strings.TrimSpace(recordType))

		if recordType == "" {
			return fmt.Errorf("record type is empty")
		}

		switch recordType {
		case "A", "AAAA", "CNAME", "TXT", "MX", "REDIRECT":
		default:
			return fmt.Errorf("unsupported record type %q", recordType)
		}

		if len(values) == 0 {
			return fmt.Errorf("record %q has no values", recordType)
		}

		for _, value := range values {
			value = strings.TrimSpace(value)

			if value == "" {
				return fmt.Errorf("record %q has empty value", recordType)
			}

			if recordType == "REDIRECT" {
				if !strings.HasPrefix(value, "https://") && !strings.HasPrefix(value, "http://") {
					return fmt.Errorf("REDIRECT target must start with http:// or https://")
				}
			}
		}
	}

	return nil
}
