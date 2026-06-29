package feature

import (
	"testing"

	"github.com/loft-sh/devpod/pkg/devcontainer/config"
)

func TestNormalizeFeatureIDRemovesVersion(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "major version",
			id:   "ghcr.io/devcontainers/features/node:1",
			want: "ghcr.io/devcontainers/features/node",
		},
		{
			name: "latest version",
			id:   "ghcr.io/devcontainers/features/node:latest",
			want: "ghcr.io/devcontainers/features/node",
		},
		{
			name: "implicit latest version",
			id:   "ghcr.io/devcontainers/features/node",
			want: "ghcr.io/devcontainers/features/node",
		},
		{
			name: "empty version",
			id:   "ghcr.io/devcontainers/features/node:",
			want: "ghcr.io/devcontainers/features/node",
		},
		{
			name: "digest",
			id:   "ghcr.io/devcontainers/features/node@sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			want: "ghcr.io/devcontainers/features/node@sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name: "invalid digest with empty algorithm separator",
			id:   "ghcr.io/devcontainers/features/node@sha256:",
			want: "ghcr.io/devcontainers/features/node@sha256:",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NormalizeFeatureID(test.id)
			if got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestNormalizeOCIFeatureReferenceTrimsEmptyTag(t *testing.T) {
	got := normalizeOCIFeatureReference("ghcr.io/devcontainers/features/node:")
	if got != "ghcr.io/devcontainers/features/node" {
		t.Fatalf("expected trailing empty tag to be removed, got %q", got)
	}
}

func TestFindContainerUsersDefaultsRemoteUserToContainerUser(t *testing.T) {
	containerUser, remoteUser := findContainerUsers(&config.ImageMetadataConfig{
		Config: []*config.ImageMetadata{
			{
				NonComposeBase: config.NonComposeBase{
					ContainerUser: "node",
				},
			},
		},
	}, "", "root")

	if containerUser != "node" {
		t.Fatalf("expected containerUser from metadata, got %q", containerUser)
	}
	if remoteUser != "node" {
		t.Fatalf("expected remoteUser to default to containerUser, got %q", remoteUser)
	}
}
