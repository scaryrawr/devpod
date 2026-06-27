package feature

import (
	"testing"

	"github.com/loft-sh/devpod/pkg/devcontainer/config"
)

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
