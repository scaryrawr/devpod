package npmconfig

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/loft-sh/devpod/pkg/util"
)

const defaultRegistry = "https://registry.npmjs.org/"

// UserRegistry returns a safe custom registry from the user's ~/.npmrc.
// Project configuration, scoped registries, credentials, and other npm settings are intentionally ignored.
func UserRegistry() (string, error) {
	homeDir, err := util.UserHomeDir()
	if err != nil {
		return "", err
	}

	file, err := os.Open(filepath.Join(homeDir, ".npmrc"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	registry := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found || !strings.EqualFold(strings.TrimSpace(key), "registry") {
			continue
		}

		registry = strings.TrimSpace(value)
		if len(registry) >= 2 {
			if first, last := registry[0], registry[len(registry)-1]; (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				registry = registry[1 : len(registry)-1]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if registry == "" || strings.TrimRight(registry, "/") == strings.TrimRight(defaultRegistry, "/") {
		return "", nil
	}
	if strings.Contains(registry, "${") {
		return "", fmt.Errorf("npm registry contains environment interpolation")
	}

	parsed, err := url.Parse(registry)
	if err != nil {
		return "", fmt.Errorf("parse npm registry: %w", err)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("npm registry must be an HTTP(S) URL")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("npm registry URL contains credentials")
	}

	return registry, nil
}
