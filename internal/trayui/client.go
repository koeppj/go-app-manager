package trayui

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/koeppj/go-app-manager/internal/app/process"
)

// APIClient is a minimal client for the service API.
type APIClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewAPIClient creates a new client with optional custom CA.
func NewAPIClient(baseURL, token, caFile string) (*APIClient, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	if caFile != "" {
		caData, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("invalid ca file")
		}
		transport.TLSClientConfig.RootCAs = pool
	}
	return &APIClient{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}, nil
}

func (c *APIClient) url(p string) string {
	return c.baseURL + path.Join("/", p)
}

func (c *APIClient) Health() error {
	req, _ := http.NewRequest(http.MethodGet, c.url("api/health"), nil)
	c.addAuth(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health status %d", resp.StatusCode)
	}
	return nil
}

func (c *APIClient) ListPrograms() ([]process.ProgramStatus, error) {
	req, _ := http.NewRequest(http.MethodGet, c.url("api/programs"), nil)
	c.addAuth(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var items []process.ProgramStatus
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *APIClient) Program(name string) (process.ProgramStatus, error) {
	req, _ := http.NewRequest(http.MethodGet, c.url(path.Join("api/programs", name)), nil)
	c.addAuth(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return process.ProgramStatus{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return process.ProgramStatus{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	var st process.ProgramStatus
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return process.ProgramStatus{}, err
	}
	return st, nil
}

func (c *APIClient) Start(name string) error   { return c.action(name, "start") }
func (c *APIClient) Stop(name string) error    { return c.action(name, "stop") }
func (c *APIClient) Restart(name string) error { return c.action(name, "restart") }

func (c *APIClient) action(name, action string) error {
	req, _ := http.NewRequest(http.MethodPost, c.url(path.Join("api/programs", name, action)), nil)
	c.addAuth(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s failed: %s", action, name, string(body))
	}
	return nil
}

func (c *APIClient) addAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}
