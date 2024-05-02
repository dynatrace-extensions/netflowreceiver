package netflowreceiver

import "fmt"

// Config represents the receiver config settings within the collector's config.yaml
type Config struct {
	Listeners []ListenerConfig `mapstructure:"listeners"`
}

type ListenerConfig struct {
	Scheme    string `mapstructure:"scheme"`
	Hostname  string `mapstructure:"hostname"`
	Port      int    `mapstructure:"port"`
	Sockets   int    `mapstructure:"sockets"`
	Workers   int    `mapstructure:"workers"`
	QueueSize int    `mapstructure:"queueSize"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	validSchemes := [3]string{"sflow", "netflow", "flow"}

	for _, listener := range cfg.Listeners {

		validScheme := false
		for _, scheme := range validSchemes {
			if listener.Scheme == scheme {
				validScheme = true
				break
			}
		}
		if !validScheme {
			return fmt.Errorf("scheme must be one of sflow, netflow, or flow")
		}

		if listener.Sockets <= 0 {
			return fmt.Errorf("sockets must be greater than 0")
		}

		if listener.Workers <= 0 {
			return fmt.Errorf("workers must be greater than 0")
		}

		if listener.QueueSize <= 0 {
			listener.QueueSize = defaultQueueSize
		}

		if listener.Port <= 0 {
			return fmt.Errorf("port must be greater than 0")
		}
	}

	return nil
}
