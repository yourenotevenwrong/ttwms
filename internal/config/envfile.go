package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// LoadFile uses the explicitly selected environment file as defaults. Process
// environment variables take precedence, including explicitly empty values.
// Values are literal: shell commands, substitutions, and escapes are not evaluated.
func LoadFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open environment file: %w", err)
	}
	defer file.Close()
	values, err := readEnv(file)
	if err != nil {
		return Config{}, err
	}
	return load(func(key string) string {
		if value, set := os.LookupEnv(key); set {
			return value
		}
		return values[key]
	})
}

func readEnv(source io.Reader) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	for number := 1; scanner.Scan(); number++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if !found || !validKey(key) {
			return nil, fmt.Errorf("environment file line %d: expected KEY=value", number)
		}
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("environment file line %d: duplicate key", number)
		}
		if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
			if len(value) < 2 || value[len(value)-1] != value[0] {
				return nil, fmt.Errorf("environment file line %d: unclosed quote", number)
			}
			value = value[1 : len(value)-1]
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read environment file: %w", err)
	}
	return values, nil
}

func validKey(key string) bool {
	if key == "" {
		return false
	}
	for i := range len(key) {
		c := key[i]
		if c != '_' && !(c >= 'A' && c <= 'Z') && !(c >= 'a' && c <= 'z') && !(i > 0 && c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
