package umbrel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

// Config holds the plugin configuration.
type Config struct{}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Provider a simple provider plugin.
type Provider struct {
	name   string
	cancel func()
}

// ContainerInfo holds the extracted container details.
type ContainerInfo struct {
	ID             string
	Names          []string
	Networks       map[string]string // network name -> ip address
	PublishedPorts []string
}

// GetContainersInfo connects to the Docker socket at unixPath (e.g. "/var/run/docker.sock")
// and returns each container's name(s), network IP(s) and published ports.
func GetContainersInfo(ctx context.Context, unixPath string) ([]ContainerInfo, error) {
	if unixPath == "" {
		unixPath = "/var/run/docker.sock"
	}

	// HTTP client that dials the unix socket
	dialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial("unix", unixPath)
	}

	transport := &http.Transport{
		DialContext: func(_ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer(_ctx, network, addr)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	baseURL := &url.URL{Scheme: "http", Host: "unix"}

	// get list of containers
	listURL := *baseURL
	listURL.Path = path.Join("/containers", "json")
	q := listURL.Query()
	q.Set("all", "1")
	listURL.RawQuery = q.Encode()

	body, err := doUnixRequest(ctx, client, listURL.String())
	if err != nil {
		return nil, err
	}

	var containers []struct {
		ID    string   `json:"Id"`
		Names []string `json:"Names"`
		Ports []struct {
			IP          string `json:"IP"`
			PrivatePort int    `json:"PrivatePort"`
			PublicPort  int    `json:"PublicPort"`
			Type        string `json:"Type"`
		} `json:"Ports"`
	}

	if err := json.Unmarshal(body, &containers); err != nil {
		return nil, err
	}

	var out []ContainerInfo
	for _, c := range containers {
		ci := ContainerInfo{
			ID:    c.ID,
			Names: c.Names,
		}

		// published ports from the list result
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				ci.PublishedPorts = append(ci.PublishedPorts, fmt.Sprintf("%s:%d->%d/%s", p.IP, p.PublicPort, p.PrivatePort, p.Type))
			} else {
				ci.PublishedPorts = append(ci.PublishedPorts, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
			}
		}

		// inspect to get network IPs
		inspURL := *baseURL
		inspURL.Path = path.Join("/containers", c.ID, "json")
		inspBody, err := doUnixRequest(ctx, client, inspURL.String())
		if err == nil {
			var insp struct {
				NetworkSettings struct {
					Networks map[string]struct {
						IPAddress string `json:"IPAddress"`
					} `json:"Networks"`
				} `json:"NetworkSettings"`
			}
			if err := json.Unmarshal(inspBody, &insp); err == nil {
				if len(insp.NetworkSettings.Networks) > 0 {
					ci.Networks = map[string]string{}
					for netName, n := range insp.NetworkSettings.Networks {
						ci.Networks[netName] = n.IPAddress
					}
				}
			}
		}

		out = append(out, ci)
	}

	return out, nil
}

func doUnixRequest(ctx context.Context, client *http.Client, rawurl string) ([]byte, error) {
	// The client dials the unix socket directly; the URL host is ignored.
	req, err := http.NewRequestWithContext(ctx, "GET", rawurl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("docker api error: %s: %s", resp.Status, string(b))
	}
	return io.ReadAll(resp.Body)
}

// New creates a new Provider plugin.
func New(ctx context.Context, config *Config, name string) (*Provider, error) {
	return &Provider{
		name: name,
	}, nil
}
