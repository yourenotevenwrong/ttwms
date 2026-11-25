package config

import (
	"strings"
	"testing"
)

func TestDefaultsAllowRunningWithoutDatabase(t *testing.T) {
	cfg, err := load(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != Development || cfg.Address != ":8080" || cfg.DatabaseURL != "" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
}

func TestDeployedEnvironmentsRequireDatabase(t *testing.T) {
	for _, environment := range []Environment{Staging, Production} {
		_, err := load(func(key string) string {
			if key == "APP_ENV" {
				return string(environment)
			}
			return ""
		})
		if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
			t.Fatalf("%s accepted missing database: %v", environment, err)
		}
	}
}

func TestLoadEnvironment(t *testing.T) {
	for _, environment := range []Environment{Development, Test, Staging, Production} {
		t.Run(string(environment), func(t *testing.T) {
			values := map[string]string{
				"APP_ENV":      string(environment),
				"APP_ADDR":     " [::1]:9090 ",
				"DATABASE_URL": "postgresql://user:password@localhost:5432/ttwms?sslmode=disable",
			}
			cfg, err := load(func(key string) string { return values[key] })
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Environment != environment || cfg.Address != "[::1]:9090" || cfg.DatabaseURL != values["DATABASE_URL"] {
				t.Fatal("configuration did not preserve the supplied settings")
			}
		})
	}
}

func TestValidAddressesAndDatabaseURLs(t *testing.T) {
	for _, address := range []string{":8080", "127.0.0.1:8080", "localhost:8080", "[::]:65535"} {
		t.Run(address, func(t *testing.T) {
			_, err := load(func(key string) string {
				if key == "APP_ADDR" {
					return address
				}
				return ""
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, databaseURL := range []string{"postgres://localhost/ttwms", "postgresql://[::1]:5432/ttwms", "postgres://user:p%40ss@postgres/ttwms?sslmode=disable"} {
		t.Run(databaseURL, func(t *testing.T) {
			_, err := load(func(key string) string {
				if key == "DATABASE_URL" {
					return databaseURL
				}
				return ""
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInvalidSettingsDoNotLeakValues(t *testing.T) {
	tests := []struct {
		key   string
		value string
	}{
		{"APP_ENV", "develpment"},
		{"APP_ADDR", "8080"},
		{"APP_ADDR", "http://localhost:8080"},
		{"APP_ADDR", "bad host:8080"},
		{"APP_ADDR", ":http"},
		{"APP_ADDR", ":0"},
		{"APP_ADDR", ":65536"},
		{"DATABASE_URL", "https://user:secret@localhost/ttwms"},
		{"DATABASE_URL", "postgres:///ttwms"},
		{"DATABASE_URL", "postgres://user:secret@localhost"},
		{"DATABASE_URL", "postgres://user:secret@localhost/"},
		{"DATABASE_URL", "postgres://user:secret@localhost:abc/ttwms"},
		{"DATABASE_URL", "postgres://user:secret@localhost:0/ttwms"},
		{"DATABASE_URL", "postgres://user:secret@localhost:65536/ttwms"},
		{"DATABASE_URL", "postgres://user:secret@localhost:/ttwms"},
		{"DATABASE_URL", "postgres://user:secret@localhost/ttwms#fragment"},
		{"DATABASE_URL", "postgres://user:secret@localhost/ttwms?sslmode=%zz"},
		{"DATABASE_URL", "postgres://user:secret%@localhost/ttwms"},
	}
	for _, tt := range tests {
		t.Run(tt.key+"/"+tt.value, func(t *testing.T) {
			_, err := load(func(key string) string {
				if key == tt.key {
					return tt.value
				}
				return ""
			})
			if err == nil || !strings.Contains(err.Error(), tt.key) {
				t.Fatalf("expected an error identifying %s, got %v", tt.key, err)
			}
			if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), tt.value) {
				t.Fatal("configuration error exposed the supplied value")
			}
		})
	}
}
