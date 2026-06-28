package agent

import "testing"

func TestDefaultAgentDownloadURLDevVersionUsesForkRelease(t *testing.T) {
	t.Setenv(EnvDevPodAgentURL, "")

	got := DefaultAgentDownloadURL()
	want := "https://github.com/scaryrawr/devpod/releases/latest/download/"
	if got != want {
		t.Fatalf("DefaultAgentDownloadURL() = %q, want %q", got, want)
	}
}

func TestDefaultAgentDownloadURLEnvOverride(t *testing.T) {
	t.Setenv(EnvDevPodAgentURL, "https://example.com/releases")

	got := DefaultAgentDownloadURL()
	want := "https://example.com/releases/"
	if got != want {
		t.Fatalf("DefaultAgentDownloadURL() = %q, want %q", got, want)
	}
}
