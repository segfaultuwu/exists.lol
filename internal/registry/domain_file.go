package registry

// DomainFile represents a domain configuration file
type DomainFile struct {
	Owner   Owner               `json:"owner"`
	Records map[string][]string `json:"records"`
}

// Owner represents the owner of a domain
type Owner struct {
	Username       string `json:"username"`
	GitHubUsername string `json:"github_username"`
	DiscordID      string `json:"discord_id"`
}

// NewDomainFile creates a new domain file with basic records
func NewDomainFile(owner Owner, records map[string][]string) DomainFile {
	return DomainFile{
		Owner:   owner,
		Records: records,
	}
}

// NewSimpleDomainFile creates a domain file with a single record type
func NewSimpleDomainFile(username, discordID, githubUsername, recordType, value string) DomainFile {
	return DomainFile{
		Owner: Owner{
			Username:       username,
			GitHubUsername: githubUsername,
			DiscordID:      discordID,
		},
		Records: map[string][]string{
			recordType: {value},
		},
	}
}

// NewDomainFileWithExtraRecords creates a domain file with additional records
func NewDomainFileWithExtraRecords(
	username string,
	discordID string,
	githubUsername string,
	recordType string,
	value string,
	extraRecords map[string]string,
) DomainFile {
	records := map[string][]string{
		recordType: {value},
	}

	for k, v := range extraRecords {
		records[k] = []string{v}
	}

	return DomainFile{
		Owner: Owner{
			Username:       username,
			GitHubUsername: githubUsername,
			DiscordID:      discordID,
		},
		Records: records,
	}
}
