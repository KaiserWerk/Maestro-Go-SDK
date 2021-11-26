package maestro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type (
	Client struct {
		baseUrl   string
		authToken string
		id        string
		client    *http.Client
	}
	ClientConfig struct {
		Timeout   time.Duration
		Transport *http.Transport
	}
	Registrant struct {
		Id      string `json:"id"`
		Address string `json:"address"`
	}

	Route string
)

const (
	authHeader = "X-Registry-Token"

	apiPrefix = "/api/v1"

	register   Route = "/register"
	deregister Route = "/deregister"
	ping       Route = "/ping"
	query      Route = "/query"
)

func New(baseUrl, token, id string, config *ClientConfig) *Client {
	httpClient := &http.Client{Timeout: 3 * time.Second}
	c := Client{baseUrl: strings.TrimSuffix(baseUrl, "/"), authToken: token, id: id}

	if config != nil {
		// set HTTP client timeout
		httpClient.Timeout = config.Timeout
		// set custom transport
		if config.Transport != nil {
			httpClient.Transport = config.Transport
		}
	}

	c.client = httpClient

	return &c
}

// Register registered the service combined with the given public address
// with the registry
func (c *Client) Register(address string) error {
	reg := Registrant{
		Id:      c.id,
		Address: address,
	}

	regJson, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.getUrl(register), bytes.NewBuffer(regJson))
	if err != nil {
		return err
	}

	c.addAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("received non-success status code (%d)", resp.StatusCode)
	}

	return nil
}

// Deregister removes the service from the registry
func (c *Client) Deregister() error {
	reg := Registrant{
		Id:      c.id,
		Address: "",
	}

	regJson, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.getUrl(deregister), bytes.NewBuffer(regJson))
	if err != nil {
		return err
	}

	c.addAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("received non-success status code (%d)", resp.StatusCode)
	}

	return nil
}

// StartPing pings the Maestro instance at the supplied interval. Should be started
// as a goroutine. Can be stopped via the context's `Cancel` function.
func (c *Client) StartPing(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := c.ping(); err != nil {
				fmt.Println("ping error: " + err.Error())
			}
		default:
		}
	}
}

func (c *Client) ping() error {
	req, err := http.NewRequest(http.MethodPut, c.getUrl(ping), nil)
	if err != nil {
		return err
	}

	c.addAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("received non-success status code (%d)", resp.StatusCode)
	}

	return nil
}

// Query queries all available IDs with addresses
func (c *Client) Query() (map[string]string, error) {
	var entries map[string]string

	req, err := http.NewRequest(http.MethodGet, c.getUrl(query), nil)
	if err != nil {
		return nil, err
	}

	c.addAuthHeader(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func (c *Client) addAuthHeader(r *http.Request) {
	r.Header.Add(authHeader, c.authToken)
}

func (c *Client) getUrl(r Route) string {
	return c.baseUrl + apiPrefix + string(r)
}


