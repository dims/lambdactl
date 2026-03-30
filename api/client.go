package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const baseURL = "https://cloud.lambdalabs.com/api/v1"

const (
	throttleBaseDelay  = 5 * time.Second
	throttleMaxDelay   = 1 * time.Minute
	throttleMaxRetries = 3
)

var (
	retrySleep  = time.Sleep
	retryJitter = func(max time.Duration) time.Duration {
		if max <= 0 {
			return 0
		}
		return time.Duration(rand.Int64N(int64(max) + 1))
	}
)

type RetryEvent struct {
	Method  string
	Path    string
	Attempt int
	Delay   time.Duration
	Err     error
}

type Client struct {
	key       string
	client    *http.Client
	retryHook func(RetryEvent)
}

func NewClient() (*Client, error) {
	// 1. Direct key from env
	key := os.Getenv("LAMBDA_API_KEY")
	if key != "" {
		return &Client{key: key, client: &http.Client{Timeout: 30 * time.Second}}, nil
	}

	// 2. Key file: explicit env var or default path
	keyFile := os.Getenv("LAMBDA_API_KEY_FILE")
	if keyFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w\n\nSet LAMBDA_API_KEY or LAMBDA_API_KEY_FILE", err)
		}
		keyFile = filepath.Join(home, ".config", "lambda", ".key")
	}
	b, err := os.ReadFile(keyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no API key found. Set one of:\n  LAMBDA_API_KEY=<key>\n  LAMBDA_API_KEY_FILE=<path>  (default: ~/.config/lambda/.key)")
		}
		return nil, fmt.Errorf("reading key file %s: %w", keyFile, err)
	}
	key = strings.TrimSpace(string(b))
	if key == "" {
		return nil, fmt.Errorf("key file %s is empty", keyFile)
	}
	return &Client{key: key, client: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) SetRetryHook(h func(RetryEvent)) {
	c.retryHook = h
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	for attempt := 0; ; attempt++ {
		var r io.Reader
		if payload != nil {
			r = bytes.NewReader(payload)
		}
		req, err := http.NewRequest(method, baseURL+path, r)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}

		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode < 400 {
			return data, nil
		}

		apiErr := newAPIError(resp, data)
		if !apiErr.IsThrottle() || attempt >= throttleMaxRetries {
			return nil, apiErr
		}

		delay := throttleDelay(apiErr, attempt)
		if c.retryHook != nil {
			c.retryHook(RetryEvent{
				Method:  method,
				Path:    path,
				Attempt: attempt + 1,
				Delay:   delay,
				Err:     apiErr,
			})
		}
		retrySleep(delay)
	}
}

func newAPIError(resp *http.Response, data []byte) *Error {
	var e struct {
		Error struct{ Message string } `json:"error"`
	}
	_ = json.Unmarshal(data, &e)
	return &Error{
		StatusCode: resp.StatusCode,
		Message:    strings.TrimSpace(e.Error.Message),
		Body:       strings.TrimSpace(string(data)),
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
	}
}

func throttleDelay(err error, attempt int) time.Duration {
	if delay, ok := RetryAfterDelay(err); ok {
		if delay > throttleMaxDelay {
			return throttleMaxDelay
		}
		return delay
	}

	delay := throttleBaseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay >= throttleMaxDelay {
			delay = throttleMaxDelay
			break
		}
	}

	jitterRange := delay / 2
	if jitterRange > 0 {
		delay += retryJitter(jitterRange)
	}
	if delay > throttleMaxDelay {
		delay = throttleMaxDelay
	}
	return delay
}

// Types

type Instance struct {
	ID               string                      `json:"id"`
	Name             string                      `json:"name,omitempty"`
	Status           string                      `json:"status"`
	IP               string                      `json:"ip,omitempty"`
	PrivateIP        string                      `json:"private_ip,omitempty"`
	Type             *InstanceType               `json:"instance_type"`
	Region           *Region                     `json:"region"`
	SSHKeyNames      []string                    `json:"ssh_key_names"`
	FileSystemNames  []string                    `json:"file_system_names"`
	FileSystemMounts []FilesystemMountEntry      `json:"file_system_mounts,omitempty"`
	Hostname         string                      `json:"hostname,omitempty"`
	JupyterToken     string                      `json:"jupyter_token,omitempty"`
	JupyterURL       string                      `json:"jupyter_url,omitempty"`
	Actions          *InstanceActionAvailability `json:"actions,omitempty"`
	Tags             []TagEntry                  `json:"tags,omitempty"`
	FirewallRulesets []FirewallRulesetEntry      `json:"firewall_rulesets,omitempty"`
}

type NamedField struct {
	Name string `json:"name"`
}

type InstanceType struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	GPUDescription string `json:"gpu_description,omitempty"`
	PriceCents     int    `json:"price_cents_per_hour"`
	Specs          Specs  `json:"specs"`
}

type Specs struct {
	VCPUs      int `json:"vcpus"`
	MemoryGiB  int `json:"memory_gib"`
	StorageGiB int `json:"storage_gib"`
	GPUs       int `json:"gpus,omitempty"`
}

type InstanceTypeEntry struct {
	Type    InstanceType `json:"instance_type"`
	Regions []Region     `json:"regions_with_capacity_available"`
}

type Region struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FilesystemMountEntry struct {
	MountPoint   string `json:"mount_point"`
	FileSystemID string `json:"file_system_id"`
}

type InstanceActionAvailability struct {
	Migrate    InstanceActionAvailabilityDetails `json:"migrate"`
	Rebuild    InstanceActionAvailabilityDetails `json:"rebuild"`
	Restart    InstanceActionAvailabilityDetails `json:"restart"`
	ColdReboot InstanceActionAvailabilityDetails `json:"cold_reboot"`
	Terminate  InstanceActionAvailabilityDetails `json:"terminate"`
}

type InstanceActionAvailabilityDetails struct {
	Available         bool   `json:"available"`
	ReasonCode        string `json:"reason_code,omitempty"`
	ReasonDescription string `json:"reason_description,omitempty"`
}

type TagEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FirewallRulesetEntry struct {
	ID string `json:"id"`
}

// API methods

func (c *Client) ListInstanceTypes() (map[string]InstanceTypeEntry, error) {
	data, err := c.do("GET", "/instance-types", nil)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data map[string]InstanceTypeEntry }
	return resp.Data, json.Unmarshal(data, &resp)
}

func (c *Client) ListInstances() ([]Instance, error) {
	data, err := c.do("GET", "/instances", nil)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data []Instance }
	return resp.Data, json.Unmarshal(data, &resp)
}

func (c *Client) GetInstance(id string) (*Instance, error) {
	data, err := c.do("GET", "/instances/"+id, nil)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data Instance }
	return &resp.Data, json.Unmarshal(data, &resp)
}

func (c *Client) Launch(gpu, sshKey, name, region string) (string, error) {
	body := map[string]any{
		"instance_type_name": gpu,
		"ssh_key_names":      []string{sshKey},
		"region_name":        region,
	}
	if name != "" {
		body["name"] = name
	}
	data, err := c.do("POST", "/instance-operations/launch", body)
	if err != nil {
		return "", err
	}
	var resp struct {
		Data struct {
			InstanceIDs []string `json:"instance_ids"`
		}
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	if len(resp.Data.InstanceIDs) == 0 {
		return "", fmt.Errorf("no instance IDs returned")
	}
	return resp.Data.InstanceIDs[0], nil
}

func (c *Client) Terminate(id string) error {
	_, err := c.do("POST", "/instance-operations/terminate", map[string]any{
		"instance_ids": []string{id},
	})
	return err
}

func (c *Client) Restart(id string) error {
	_, err := c.do("POST", "/instance-operations/restart", map[string]any{
		"instance_ids": []string{id},
	})
	return err
}

func (c *Client) RenameInstance(id, name string) (*Instance, error) {
	data, err := c.do("POST", "/instances/"+id, map[string]any{
		"name": name,
	})
	if err != nil {
		return nil, err
	}
	var resp struct{ Data Instance }
	return &resp.Data, json.Unmarshal(data, &resp)
}

// SSH keys

type SSHKey struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"`
}

func (c *Client) ListSSHKeys() ([]SSHKey, error) {
	data, err := c.do("GET", "/ssh-keys", nil)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data []SSHKey }
	return resp.Data, json.Unmarshal(data, &resp)
}

func (c *Client) AddSSHKey(name, publicKey string) (*SSHKey, error) {
	body := map[string]string{"name": name}
	if publicKey != "" {
		body["public_key"] = publicKey
	}
	data, err := c.do("POST", "/ssh-keys", body)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data SSHKey }
	return &resp.Data, json.Unmarshal(data, &resp)
}

func (c *Client) DeleteSSHKey(id string) error {
	_, err := c.do("DELETE", "/ssh-keys/"+id, nil)
	return err
}

// Images

type Image struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Family      string     `json:"family"`
	Version     string     `json:"version"`
	Region      NamedField `json:"region"`
}

func (c *Client) ListImages() ([]Image, error) {
	data, err := c.do("GET", "/images", nil)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data []Image }
	return resp.Data, json.Unmarshal(data, &resp)
}

// Firewall rules

type FirewallRule struct {
	Protocol      string `json:"protocol"`
	PortRange     [2]int `json:"port_range"`
	SourceNetwork string `json:"source_network"`
	Description   string `json:"description"`
}

func (c *Client) ListFirewallRules() ([]FirewallRule, error) {
	data, err := c.do("GET", "/firewall-rules", nil)
	if err != nil {
		return nil, err
	}
	var resp struct{ Data []FirewallRule }
	return resp.Data, json.Unmarshal(data, &resp)
}
