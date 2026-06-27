//go:build !windows

package config

import (
	"strings"
	"testing"
)

func TestPatchEnvVarsPreservesValuesContainingEquals(t *testing.T) {
	env := patchEnvVars([]string{
		"PATH=/usr/local/bin:/usr/bin",
		"TOKEN=prefix=value",
	}, map[string]string{
		"HOME": "/home/vscode",
	})

	envMap := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("expected key=value entry, got %q", entry)
		}
		envMap[key] = value
	}

	if envMap["TOKEN"] != "prefix=value" {
		t.Fatalf("expected TOKEN to preserve equals, got %q", envMap["TOKEN"])
	}
	if envMap["HOME"] != "/home/vscode" {
		t.Fatalf("expected HOME patch, got %q", envMap["HOME"])
	}
}
