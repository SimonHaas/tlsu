package umbrel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/traefik/genconf/dynamic"
)

// Config holds the plugin configuration.
type Config struct{}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// Provider a simple provider plugin.
type Provider struct {
	name         string
	pollInterval time.Duration
	cancel       func()
}

// ContainerInfo holds the extracted container details.
type ContainerInfo struct {
	ID             string
	Names          []string
	Networks       map[string]string // network name -> ip address
	PublishedPorts []PublishedPort
}

// PublishedPort describes a published port mapping from Docker.
type PublishedPort struct {
	IP          string
	PublicPort  int
	PrivatePort int
	Proto       string
}

// GetContainersInfo connects to the Docker socket at unixPath (e.g. "/var/run/docker.sock")
// and returns each container's name(s), network IP(s) and published ports.
func GetContainersInfo(ctx context.Context, unixPath string) ([]ContainerInfo, error) {
	fmt.Println("umbrel - GetContainersInfo")
	
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
			pp := PublishedPort{
				IP:          p.IP,
				PublicPort:  p.PublicPort,
				PrivatePort: p.PrivatePort,
				Proto:       p.Type,
			}
			ci.PublishedPorts = append(ci.PublishedPorts, pp)
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

		// apply filters: only include containers attached to umbrel_main_network
		// and whose name ends with _app_proxy_1
		networkFilter := "umbrel_main_network"
		nameSuffix := "_app_proxy_1"

		if ci.Networks == nil {
			continue
		}
		if _, ok := ci.Networks[networkFilter]; !ok {
			continue
		}

		nameMatch := false
		for _, n := range ci.Names {
			if strings.HasSuffix(n, nameSuffix) {
				nameMatch = true
				break
			}
		}
		if !nameMatch {
			continue
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
	fmt.Println("umbrel - New")

	// default poll interval
	pi := 10 * time.Second

	return &Provider{
		name:         name,
		pollInterval: pi,
	}, nil
}

// Init the provider.
func (p *Provider) Init() error {
	fmt.Println("umbrel - Init")
	
	if p.pollInterval <= 0 {
		return fmt.Errorf("poll interval must be greater than 0")
	}
	return nil
}

// Provide creates and sends dynamic configuration to Traefik.
func (p *Provider) Provide(cfgChan chan<- json.Marshaler) error {
	fmt.Println("umbrel - Provide")
	
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Print(err)
			}
		}()

		p.loadConfiguration(ctx, cfgChan)
	}()

	return nil
}

func (p *Provider) loadConfiguration(ctx context.Context, cfgChan chan<- json.Marshaler) {
	fmt.Println("umbrel - loadConfiguration")

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			infos, err := GetContainersInfo(ctx, "")
			if err != nil {
				log.Print(err)
				continue
			}

			configuration := &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     make(map[string]*dynamic.Router),
					Middlewares: make(map[string]*dynamic.Middleware),
					Services:    make(map[string]*dynamic.Service),
				},
			}

			networkFilter := "umbrel_main_network"

			for _, ci := range infos {
				// determine container IPv4 on the umbrel network
				ip := ci.Networks[networkFilter]
				if ip == "" {
					continue
				}
				if net.ParseIP(ip) == nil || net.ParseIP(ip).To4() == nil {
					continue
				}

				// find a suitable private port (tcp)
				var port int
				for _, pp := range ci.PublishedPorts {
					if pp.PrivatePort > 0 && strings.ToLower(pp.Proto) == "tcp" {
						port = pp.PrivatePort
						break
					}
				}
				if port == 0 {
					continue
				}

				svcName := fmt.Sprintf("umbrel-service-%s", ci.ID[:12])
				url := fmt.Sprintf("http://%s:%d", ip, port)

				configuration.HTTP.Services[svcName] = &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Servers: []dynamic.Server{
							{URL: url},
						},
						PassHostHeader: boolPtr(true),
					},
				}
				// create a host-based router using the first part of the container name
				// e.g. "/metube_app_proxy_1" -> metube.umbrel.simonhaas.eu
				var firstName string
				if len(ci.Names) > 0 {
					firstName = strings.TrimPrefix(ci.Names[0], "/")
				}
				if firstName != "" {
					parts := strings.Split(firstName, "_")
					if len(parts) > 0 {
						host := fmt.Sprintf("%s.umbrel.simonhaas.eu", parts[0])
						routerName := fmt.Sprintf("umbrel-router-%s", ci.ID[:12])
						configuration.HTTP.Routers[routerName] = &dynamic.Router{
							EntryPoints: []string{"websecure"},
							Service:     svcName,
							Rule:        fmt.Sprintf("Host(`%s`)", host),
						}
					}
				}
			}

			cfgChan <- &dynamic.JSONPayload{Configuration: configuration}

		case <-ctx.Done():
			return
		}
	}
}

// Stop to stop the provider and the related go routines.
func (p *Provider) Stop() error {
	fmt.Println("umbrel - Stop")

	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func boolPtr(v bool) *bool { return &v }
