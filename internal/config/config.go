package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Environment string

const (
	Development Environment = "development"
	Test        Environment = "test"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

type Config struct {
	Environment Environment
	Address     string
	DatabaseURL string
}

// Load reads and validates process environment variables. It does not connect to
// the database; an empty database URL is allowed while persistence is optional.
func Load() (Config, error) {
	return load(os.Getenv)
}

func load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Environment: Environment(strings.TrimSpace(getenv("APP_ENV"))),
		Address:     strings.TrimSpace(getenv("APP_ADDR")),
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
	}
	if cfg.Environment == "" {
		cfg.Environment = Development
	}
	switch cfg.Environment {
	case Development, Test, Staging, Production:
	default:
		return Config{}, errors.New("APP_ENV must be development, test, staging, or production")
	}
	if cfg.Address == "" {
		cfg.Address = ":8080"
	}
	host, port, err := net.SplitHostPort(cfg.Address)
	if err != nil || strings.ContainsAny(host, " \t\r\n/?#@") || !validPort(port) {
		return Config{}, errors.New("APP_ADDR must be a host:port address with a port from 1 to 65535")
	}
	if cfg.DatabaseURL != "" {
		u, err := url.Parse(cfg.DatabaseURL)
		if err != nil {
			// URL parsing errors may contain credentials. Never return the input.
			return Config{}, errors.New("DATABASE_URL must be a valid PostgreSQL URL")
		}
		if (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.Path == "" || u.Path == "/" || u.Fragment != "" || strings.HasSuffix(u.Host, ":") {
			return Config{}, errors.New("DATABASE_URL must use postgres:// or postgresql:// with a host and database name and no fragment")
		}
		if port := u.Port(); port != "" && !validPort(port) {
			return Config{}, errors.New("DATABASE_URL port must be from 1 to 65535")
		}
		if _, err := url.ParseQuery(u.RawQuery); err != nil {
			return Config{}, errors.New("DATABASE_URL must have valid query parameters")
		}
	}
	return cfg, nil
}

func validPort(port string) bool {
	if port == "" || strings.IndexFunc(port, func(r rune) bool { return r < '0' || r > '9' }) >= 0 {
		return false
	}
	n, err := strconv.Atoi(port)
	return err == nil && n >= 1 && n <= 65535
}
