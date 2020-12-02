package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// config is a stuct with all config values. See `runtime/config/config.toml`
// for more information about these values.
type config struct {
	Listen string

	MailHost string
	MailPort int
	MailUser string
	MailPass string
}

// parseConfig parses a toml config.
func parseConfig() (*config, error) {
	c := &config{}

	if _, err := toml.DecodeFile("/etc/wimpel/config.toml", c); err != nil {
		return nil, fmt.Errorf("config %s: %s", "/etc/wimpel/config.toml", err)
	}

	return c, nil
}
