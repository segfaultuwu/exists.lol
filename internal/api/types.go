package api

type DomainResponse struct {
	Subdomain string              `json:"subdomain"`
	FQDN      string              `json:"fqdn"`
	Records   map[string][]string `json:"records"`
	Owner     Owner               `json:"owner"`
	Status    string              `json:"status"`
}

type Owner struct {
	Username string `json:"username,omitempty"`
	Discord  string `json:"discord,omitempty"`
	GitHub   string `json:"github,omitempty"`
}

type StatsResponse struct {
	DomainsTotal int `json:"domains_total"`
	Active       int `json:"active"`
	Suspended    int `json:"suspended"`
}

type StatusResponse struct {
	OK       bool   `json:"ok"`
	Service  string `json:"service"`
	Registry string `json:"registry"`
}

type ValidateRequest struct {
	Subdomain string `json:"subdomain"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

type ValidateResponse struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

type CreateDomainRequest struct {
	DiscordID      string              `json:"discord_id"`
	Username       string              `json:"username"`
	GitHubUsername string              `json:"github_username"`
	Subdomain      string              `json:"subdomain"`
	Records        map[string][]string `json:"records"`
}

type CreateDomainResponse struct {
	OK        bool   `json:"ok"`
	Subdomain string `json:"subdomain"`
	FQDN      string `json:"fqdn"`
	Message   string `json:"message"`
}

type ReloadRegistryResponse struct {
	OK      bool `json:"ok"`
	Domains int  `json:"domains"`
}
