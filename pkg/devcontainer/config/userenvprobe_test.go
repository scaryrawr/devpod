package config

import (
	"testing"

	"github.com/loft-sh/devpod/pkg/log"
)

func TestParseUserEnvOutputPreservesValuesContainingEquals(t *testing.T) {
	env, err := parseUserEnvOutput([]byte("TOKEN=prefix=value\x00PWD=/workspace\x00NO_EQUALS\x00"), '\x00', log.Discard)
	if err != nil {
		t.Fatalf("parse user env output: %v", err)
	}

	if env["TOKEN"] != "prefix=value" {
		t.Fatalf("expected TOKEN to preserve equals, got %q", env["TOKEN"])
	}
	if env["PWD"] != "/workspace" {
		t.Fatalf("expected PWD to be parsed before doProbe filters it, got %q", env["PWD"])
	}
	if _, ok := env["NO_EQUALS"]; ok {
		t.Fatalf("expected invalid env entry to be ignored")
	}
}
