package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDevPodSourceRootFrom(t *testing.T) {
	tempDir := t.TempDir()
	root := filepath.Join(tempDir, "repo")
	nested := filepath.Join(root, "pkg", "agent")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/loft-sh/devpod\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	got, err := findDevPodSourceRootFrom(nested)
	if err != nil {
		t.Fatalf("find source root: %v", err)
	}
	if got != root {
		t.Fatalf("findDevPodSourceRootFrom() = %q, want %q", got, root)
	}
}

func TestFindDevPodSourceRootFromMissing(t *testing.T) {
	got, err := findDevPodSourceRootFrom(t.TempDir())
	if err != nil {
		t.Fatalf("find source root: %v", err)
	}
	if got != "" {
		t.Fatalf("findDevPodSourceRootFrom() = %q, want empty", got)
	}
}
