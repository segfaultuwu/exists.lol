package validate

import (
	"fmt"
	"regexp"
	"strings"
)

var allowedSubdomain = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var reserved = map[string]bool{
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

func Request(subdomain, recordType, value string) error {
	subdomain = strings.TrimSpace(subdomain)
	recordType = strings.ToUpper(strings.TrimSpace(recordType))
	value = strings.TrimSpace(value)

	if subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}

	if len(subdomain) > 63 {
		return fmt.Errorf("subdomain is too long")
	}

	if !allowedSubdomain.MatchString(subdomain) {
		return fmt.Errorf("subdomain can only contain lowercase letters, numbers and dashes")
	}

	if strings.HasPrefix(subdomain, "-") || strings.HasSuffix(subdomain, "-") {
		return fmt.Errorf("subdomain cannot start or end with dash")
	}

	if reserved[subdomain] {
		return fmt.Errorf("this subdomain is reserved")
	}

	switch recordType {
	case "A", "AAAA", "CNAME", "TXT", "MX":
	default:
		return fmt.Errorf("unsupported record type")
	}

	if value == "" {
		return fmt.Errorf("record value is required")
	}

	if strings.Contains(value, "*") {
		return fmt.Errorf("wildcards are not allowed")
	}

	return nil
}
