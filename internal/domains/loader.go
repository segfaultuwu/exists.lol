package domains

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Config struct {
	Owner   Owner               `json:"owner"`
	Records map[string]RawValue `json:"records"`
}

type Owner struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RawValue []string

type Domain struct {
	Subdomain string
	Config    Config
}

var allowedSubdomain = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

var reservedSubdomains = map[string]bool{
	"www":       true,
	"mail":      true,
	"api":       true,
	"admin":     true,
	"root":      true,
	"support":   true,
	"ns1":       true,
	"ns2":       true,
	"ftp":       true,
	"dashboard": true,
	"status":    true,
	"cdn":       true,
	"assets":    true,
	"login":     true,
	"auth":      true,
	"account":   true,
	"accounts":  true,
	"billing":   true,
	"security":  true,
}

var allowedRecords = map[string]bool{
	"A":        true,
	"AAAA":     true,
	"CNAME":    true,
	"TXT":      true,
	"MX":       true,
	"REDIRECT": true,
}

func (r *RawValue) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*r = []string{single}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*r = many
		return nil
	}

	return fmt.Errorf("record value must be string or string array")
}

// a small helper
func IsWildcard(domain string) bool {
	if strings.Contains(domain, "*") {
		return true
	} else {
		return false
	}
}

func Load(dir string) ([]Domain, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type result struct {
		domain Domain
		err    error
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".json") {
			continue
		}

		if IsWildcard(name) {
			continue
		}

		files = append(files, name)
	}

	results := make(chan result, len(files))

	var wg sync.WaitGroup

	for _, file := range files {
		wg.Go(func() {
			domain, err := loadOne(dir, file)
			results <- result{
				domain: domain,
				err:    err,
			}
		})
	}

	wg.Wait()
	close(results)

	var domains []Domain
	var errors []string

	for res := range results {
		if res.err != nil {
			errors = append(errors, res.err.Error())
			continue
		}

		domains = append(domains, res.domain)
	}

	if len(errors) > 0 {
		sort.Strings(errors)
		return nil, fmt.Errorf("validation failed:\n%s", strings.Join(errors, "\n"))
	}

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Subdomain < domains[j].Subdomain
	})

	return domains, nil
}

func loadOne(dir string, name string) (Domain, error) {
	subdomain := strings.TrimSuffix(name, ".json")

	if err := validateSubdomain(subdomain); err != nil {
		return Domain{}, fmt.Errorf("%s: %w", name, err)
	}

	fullPath := filepath.Join(dir, name)

	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return Domain{}, fmt.Errorf("%s: %w", name, err)
	}

	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Domain{}, fmt.Errorf("%s: %w", name, err)
	}

	if err := validateConfig(subdomain, cfg); err != nil {
		return Domain{}, fmt.Errorf("%s: %w", name, err)
	}

	return Domain{
		Subdomain: subdomain,
		Config:    cfg,
	}, nil
}

func validateSubdomain(name string) error {
	if name == "" {
		return fmt.Errorf("empty subdomain")
	}

	if len(name) > 63 {
		return fmt.Errorf("subdomain too long: %s", name)
	}

	if !allowedSubdomain.MatchString(name) {
		return fmt.Errorf("invalid subdomain: %s", name)
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("subdomain cannot start or end with dash: %s", name)
	}

	if reservedSubdomains[name] {
		return fmt.Errorf("reserved subdomain: %s", name)
	}

	return nil
}

func validateConfig(subdomain string, cfg Config) error {
	if cfg.Owner.Username == "" {
		return fmt.Errorf("owner.username is required for %s", subdomain)
	}

	if len(cfg.Records) == 0 {
		return fmt.Errorf("records are required")
	}

	if _, hasCNAME := cfg.Records["CNAME"]; hasCNAME && len(cfg.Records) > 1 {
		return fmt.Errorf("CNAME cannot be mixed with other record types")
	}

	for recordType, values := range cfg.Records {
		if !allowedRecords[recordType] {
			return fmt.Errorf("record type not allowed: %s", recordType)
		}

		if len(values) == 0 {
			return fmt.Errorf("empty value list for %s", recordType)
		}

		for _, value := range values {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("empty value for %s", recordType)
			}

			if strings.Contains(value, "*") {
				return fmt.Errorf("wildcards are not allowed")
			}
		}
	}

	return nil
}
