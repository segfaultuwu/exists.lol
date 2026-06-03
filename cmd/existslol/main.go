package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/segfaultuwu/exists.lol/internal/cloudflare"
	"github.com/segfaultuwu/exists.lol/internal/domains"
)

func main() {
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")
	rootDomain := os.Getenv("ROOT_DOMAIN")

	if token == "" {
		die("missing CLOUDFLARE_API_TOKEN")
	}

	if zoneID == "" {
		die("missing CLOUDFLARE_ZONE_ID")
	}

	if rootDomain == "" {
		die("missing ROOT_DOMAIN")
	}

	rootDomain = strings.TrimSuffix(rootDomain, ".")

	loaded, err := domains.Load("domains")
	if err != nil {
		die(err.Error())
	}

	cf := cloudflare.New(token, zoneID)

	existing, err := cf.ListRecords()
	if err != nil {
		die(err.Error())
	}

	for _, domain := range loaded {
		fqdn := domain.Subdomain + "." + rootDomain

		for recordType, values := range domain.Config.Records {
			for _, value := range values {
				value = strings.TrimSpace(value)

				record := cloudflare.DNSRecord{
					Type:    recordType,
					Name:    fqdn,
					Content: value,
				}

				current := findRecord(existing, fqdn, recordType)

				if current == nil {
					fmt.Printf("[create] %s %s %s\n", fqdn, recordType, value)

					if err := cf.CreateRecord(record); err != nil {
						die(err.Error())
					}

					continue
				}

				if current.Content == value {
					fmt.Printf("[skip] %s %s %s\n", fqdn, recordType, value)
					continue
				}

				fmt.Printf("[update] %s %s %s -> %s\n", fqdn, recordType, current.Content, value)

				if err := cf.UpdateRecord(current.ID, record); err != nil {
					die(err.Error())
				}
			}
		}
	}
}

func findRecord(records []cloudflare.DNSRecord, name, recordType string) *cloudflare.DNSRecord {
	for i := range records {
		record := &records[i]

		if record.Name == name && record.Type == recordType {
			return record
		}
	}

	return nil
}

func die(msg string) {
	fmt.Fprintln(os.Stderr, "error:", msg)
	os.Exit(1)
}
