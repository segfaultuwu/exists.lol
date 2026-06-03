package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	Token  string
	ZoneID string
	HTTP   *http.Client
}

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type listResponse struct {
	Success bool        `json:"success"`
	Errors  []CFError   `json:"errors"`
	Result  []DNSRecord `json:"result"`
}

type createResponse struct {
	Success bool      `json:"success"`
	Errors  []CFError `json:"errors"`
	Result  DNSRecord `json:"result"`
}

type CFError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(token, zoneID string) *Client {
	return &Client{
		Token:  token,
		ZoneID: zoneID,
		HTTP:   http.DefaultClient,
	}
}

func (c *Client) ListRecords() ([]DNSRecord, error) {
	endpoint := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/zones/%s/dns_records?per_page=5000",
		url.PathEscape(c.ZoneID),
	)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	c.auth(req)

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parsed listResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	if !parsed.Success {
		return nil, fmt.Errorf("cloudflare list failed: %s", formatErrors(parsed.Errors))
	}

	return parsed.Result, nil
}

func (c *Client) UpdateRecord(id string, record DNSRecord) error {
	endpoint := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
		url.PathEscape(c.ZoneID),
		url.PathEscape(id),
	)

	payload := map[string]any{
		"type":    record.Type,
		"name":    record.Name,
		"content": record.Content,
		"ttl":     1,
		"proxied": canProxy(record.Type),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}

	c.auth(req)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var parsed createResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}

	if !parsed.Success {
		return fmt.Errorf("cloudflare update failed: %s", formatErrors(parsed.Errors))
	}

	return nil
}

func canProxy(recordType string) bool {
	switch recordType {
	case "A", "AAAA", "CNAME":
		return true
	default:
		return false
	}
}

func (c *Client) DeleteRecord(id string) error {
	endpoint := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
		url.PathEscape(c.ZoneID),
		url.PathEscape(id),
	)

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	c.auth(req)

	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var parsed createResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}

	if !parsed.Success {
		return fmt.Errorf("cloudflare delete failed: %s", formatErrors(parsed.Errors))
	}

	return nil
}

func (c *Client) CreateRecord(record DNSRecord) error {
	endpoint := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/zones/%s/dns_records",
		url.PathEscape(c.ZoneID),
	)

	payload := map[string]any{
		"type":    record.Type,
		"name":    record.Name,
		"content": record.Content,
		"ttl":     1,
		"proxied": canProxy(record.Type),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}

	c.auth(req)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var parsed createResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}

	if !parsed.Success {
		return fmt.Errorf("cloudflare create failed: %s", formatErrors(parsed.Errors))
	}

	return nil
}

func (c *Client) auth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.Token)
}

func formatErrors(errors []CFError) string {
	if len(errors) == 0 {
		return "unknown error"
	}

	msg := ""

	for i, err := range errors {
		if i > 0 {
			msg += "; "
		}

		msg += fmt.Sprintf("%d: %s", err.Code, err.Message)
	}

	return msg
}
