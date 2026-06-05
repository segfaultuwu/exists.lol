package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: baseURL,
		token:   token,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Owner struct {
	Username string `json:"username,omitempty"`
	Discord  string `json:"discord,omitempty"`
	GitHub   string `json:"github,omitempty"`
}

type DomainResponse struct {
	Subdomain string              `json:"subdomain"`
	FQDN      string              `json:"fqdn"`
	Records   map[string][]string `json:"records"`
	Owner     Owner               `json:"owner"`
	Status    string              `json:"status"`
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

type APIError struct {
	Error string `json:"error"`
}

func (c *Client) GetDomain(ctx context.Context, subdomain string) (*DomainResponse, error) {
	var out DomainResponse

	if err := c.getJSON(ctx, "/api/domains/"+subdomain, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) ListDomains(ctx context.Context) ([]DomainResponse, error) {
	var out []DomainResponse

	if err := c.getJSON(ctx, "/api/domains", &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *Client) CreateDomain(ctx context.Context, req CreateDomainRequest) (*CreateDomainResponse, error) {
	var out CreateDomainResponse

	if err := c.postJSON(ctx, "/api/domains", req, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) ReloadRegistry(ctx context.Context) (*ReloadRegistryResponse, error) {
	var out ReloadRegistryResponse

	if err := c.postJSON(ctx, "/api/registry/reload", map[string]any{}, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	return c.do(req, out)
}

func (c *Client) postJSON(ctx context.Context, path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)

		if apiErr.Error != "" {
			return fmt.Errorf("%s", apiErr.Error)
		}

		return fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
