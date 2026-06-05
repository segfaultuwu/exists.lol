package registry

import (
	"fmt"
	"os/exec"
	"strings"
)

func normalizeRecords(records map[string][]string) map[string][]string {
	out := make(map[string][]string, len(records))

	for recordType, values := range records {
		recordType = strings.ToUpper(strings.TrimSpace(recordType))
		if recordType == "" {
			continue
		}

		cleanValues := make([]string, 0, len(values))

		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}

			cleanValues = append(cleanValues, value)
		}

		if len(cleanValues) > 0 {
			out[recordType] = cleanValues
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
