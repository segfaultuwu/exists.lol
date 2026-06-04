package registry

type DomainFile struct {
	Owner   Owner               `json:"owner"`
	Records map[string][]string `json:"records"`
}

type Owner struct {
	Username       string `json:"username"`
	GitHubUsername string `json:"github_username"`
	DiscordID      string `json:"discord_id"`
}

func NewDomainFile(discordUsername, discordID, githubUsername, recordType, value string) DomainFile {
	return NewDomainFileWithExtraRecords(
		discordUsername,
		discordID,
		githubUsername,
		recordType,
		value,
		nil,
	)
}

func NewDomainFileWithExtraRecords(
	discordUsername string,
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
			Username:       discordUsername,
			GitHubUsername: githubUsername,
			DiscordID:      discordID,
		},
		Records: records,
	}
}
