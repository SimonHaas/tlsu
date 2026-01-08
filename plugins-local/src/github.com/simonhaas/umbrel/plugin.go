package umbrel

import (
    "context"
)

// Config holds the plugin configuration.
type Config struct {}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
    return &Config{}
}

// Provider a simple provider plugin.
type Provider struct {
	name         string

	cancel func()
}

// New creates a new Provider plugin.
func New(ctx context.Context, config *Config, name string) (*Provider, error) {
	return &Provider{
		name:         name,
	}, nil
}
