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
		BaseUrl   string
		AuthToken string
		Id        string
		Client    *http.Client
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
	query      Route = "/query?id=%s"
)

func New(baseUrl, token, id string, config *ClientConfig) *Client {
	httpClient := &http.Client{Timeout: 3 * time.Second}
	c := Client{BaseUrl: strings.TrimSuffix(baseUrl, "/"), AuthToken: token, Id: id}

	if config != nil {
		// set HTTP client timeout
		if config.Timeout > 10*time.Millisecond {
			httpClient.Timeout = config.Timeout
		}
		// set custom transport
		if config.Transport != nil {
			httpClient.Transport = config.Transport
		}
	}

	c.Client = httpClient

	return &c
}

// Register registered the service combined with the given public address
// with the registry
func (c *Client) Register(address string) error {
	reg := Registrant{
		Id:      c.Id,
		Address: address,
	}

	regJson, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.getUrl(register), bytes.NewBuffer(regJson))
	if err != nil {
		return err
	}

	c.addAuthHeader(req)

	resp, err := c.Client.Do(req)
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
		Id:      c.Id,
		Address: "",
	}

	regJson, err := json.Marshal(reg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, c.getUrl(deregister), bytes.NewBuffer(regJson))
	if err != nil {
		return err
	}

	c.addAuthHeader(req)

	resp, err := c.Client.Do(req)
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
// as a goroutine. Can be stopped via the context's Cancel function.
func (c *Client) StartPing(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := c.Ping(); err != nil {
				fmt.Println("ping error: " + err.Error())
			}
		default:
		}
	}
}

func (c *Client) Ping() error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s?id=%s", c.getUrl(ping), c.Id), nil)
	if err != nil {
		return err
	}

	c.addAuthHeader(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("received non-success status code (%d)", resp.StatusCode)
	}

	return nil
}

// Query queries Maestro for the info on the specified ID
func (c *Client) Query(id string) (Registrant, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(c.getUrl(query), id), nil)
	if err != nil {
		return Registrant{}, fmt.Errorf("could not create request: %s", err.Error())
	}
	c.addAuthHeader(req)

	resp, err := c.Client.Do(req)
	if err != nil {
		return Registrant{}, fmt.Errorf("could not execute query request: %s", err.Error())
	}
	defer resp.Body.Close()

	var entry Registrant
	err = json.NewDecoder(resp.Body).Decode(&entry)
	if err != nil {
		return Registrant{}, fmt.Errorf("could not decode JSON: %s", err.Error())
	}

	return entry, nil
}

func (c *Client) addAuthHeader(r *http.Request) {
	r.Header.Add(authHeader, c.AuthToken)
}

func (c *Client) getUrl(r Route) string {
	return c.BaseUrl + apiPrefix + string(r)
}
