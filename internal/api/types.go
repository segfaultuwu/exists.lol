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
