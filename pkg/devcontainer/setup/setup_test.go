package setup

import "testing"

func TestShouldChownWorkspaceSkipsRoot(t *testing.T) {
	if shouldChownWorkspace("/") {
		t.Fatalf("expected root workspace folder to skip recursive chown")
	}
}

func TestShouldChownWorkspaceAllowsNonRoot(t *testing.T) {
	if !shouldChownWorkspace("/workspaces/project") {
		t.Fatalf("expected non-root workspace folder to allow recursive chown")
	}
}
