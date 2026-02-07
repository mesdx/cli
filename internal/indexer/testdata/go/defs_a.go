package sample

import "fmt"

// Config holds application configuration.
// It supports multiple environments.
type Config struct {
	Host     string
	Port     int
	LogLevel string
}

// Validate checks if the config is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("port must be positive")
	}
	return nil
}

// Formatter formats output.
type Formatter interface {
	Format(data string) string
	Reset()
}

// DefaultPort is the fallback port.
const DefaultPort = 8080

// AppName is the application name.
var AppName = "myapp"

// NewConfig creates a Config with defaults.
func NewConfig(host string) *Config {
	return &Config{
		Host:     host,
		Port:     DefaultPort,
		LogLevel: "info",
	}
}

type (
	// Endpoint describes an API endpoint.
	Endpoint struct {
		Path   string
		Method string
	}

	// Handler processes requests.
	Handler func(req string) string
)
