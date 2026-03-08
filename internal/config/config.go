package config

import (
	"fmt"
	"os"
	"time"
)

// Adapter holds the registration and settings for a single adapter binary.
type Adapter struct {
	Bin     string            `mapstructure:"bin"`
	Timeout time.Duration     `mapstructure:"timeout"`
	Num     int               `mapstructure:"num"`
	Config  map[string]string `mapstructure:"config"`
}

// Routing controls how qry selects and combines adapters.
type Routing struct {
	Mode     string   `mapstructure:"mode"`     // "first" or "merge"
	Pool     []string `mapstructure:"pool"`     // adapters actively used for queries
	Fallback []string `mapstructure:"fallback"` // "first" mode only
}

// Defaults holds global fallback values applied when not set per-adapter.
type Defaults struct {
	Num     int           `mapstructure:"num"`
	Timeout time.Duration `mapstructure:"timeout"`
}

// Config is the fully resolved configuration for a qry invocation.
type Config struct {
	Defaults Defaults           `mapstructure:"defaults"`
	Routing  Routing            `mapstructure:"routing"`
	Adapters map[string]Adapter `mapstructure:"adapters"`
}

// Validate checks the config for required fields and consistency.
func (c *Config) Validate() error {
	if len(c.Routing.Pool) == 0 {
		return fmt.Errorf("routing.pool must contain at least one adapter")
	}
	if c.Routing.Mode != "first" && c.Routing.Mode != "merge" {
		return fmt.Errorf("routing.mode must be \"first\" or \"merge\", got %q", c.Routing.Mode)
	}
	for _, name := range append(c.Routing.Pool, c.Routing.Fallback...) {
		adapter, ok := c.Adapters[name]
		if !ok {
			return fmt.Errorf("adapter %q referenced in routing but not declared in [adapters]", name)
		}
		if adapter.Bin == "" {
			return fmt.Errorf("adapter %q is missing required field: bin", name)
		}
		if _, err := os.Stat(adapter.Bin); err != nil {
			return fmt.Errorf("adapter %q binary not found at %q: %w", name, adapter.Bin, err)
		}
	}
	return nil
}

// ResolvedAdapter returns the adapter config for the given name with defaults applied.
func (c *Config) ResolvedAdapter(name string) (Adapter, error) {
	a, ok := c.Adapters[name]
	if !ok {
		return Adapter{}, fmt.Errorf("adapter %q not found in config", name)
	}
	if a.Timeout == 0 {
		a.Timeout = c.Defaults.Timeout
	}
	if a.Num == 0 {
		a.Num = c.Defaults.Num
	}
	return a, nil
}
