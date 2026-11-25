package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnvFileValuesAreLiteral(t *testing.T) {
	values, err := readEnv(strings.NewReader("# comment\nAPP_ENV = development\nAPP_ADDR='127.0.0.1:9090'\nPASSWORD=\"$(touch should-not-exist)=$HOME\"\nEMPTY=\n"))
	if err != nil {
		t.Fatal(err)
	}
	if values["APP_ADDR"] != "127.0.0.1:9090" || values["PASSWORD"] != "$(touch should-not-exist)=$HOME" || values["EMPTY"] != "" {
		t.Fatal("file values were not preserved literally")
	}
}

func TestInvalidEnvFileDoesNotExposeValues(t *testing.T) {
	for _, input := range []string{"DATABASE_URL='private-secret", "BAD KEY=private-secret", "private-secret", "KEY=private-secret\nKEY=another-secret"} {
		_, err := readEnv(strings.NewReader(input))
		if err == nil || !strings.Contains(err.Error(), "line") || strings.Contains(err.Error(), "secret") {
			t.Fatalf("expected safe line-specific error, got %v", err)
		}
	}
}

func TestLoadFileProcessEnvironmentTakesPrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("APP_ENV=development\nAPP_ADDR=127.0.0.1:9090\nDATABASE_URL=postgres://localhost/file_database\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_ENV", "test")
	t.Setenv("DATABASE_URL", "")
	// Ensure the address is read from the file regardless of the caller's shell.
	address, set := os.LookupEnv("APP_ADDR")
	if err := os.Unsetenv("APP_ADDR"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if set {
			_ = os.Setenv("APP_ADDR", address)
		}
	})
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Environment != Test || cfg.Address != "127.0.0.1:9090" || cfg.DatabaseURL != "" {
		t.Fatal("process overrides or file defaults were not respected")
	}
}

func TestMissingExplicitEnvFileFails(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing requested environment file was silently ignored")
	}
}
