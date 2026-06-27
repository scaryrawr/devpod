package config

import (
	"testing"

	"github.com/loft-sh/devpod/pkg/types"
)

func TestMergeConfigurationAppliesWaitForDefault(t *testing.T) {
	merged, err := MergeConfiguration(&DevContainerConfig{}, nil)
	if err != nil {
		t.Fatalf("MergeConfiguration() error = %v", err)
	}

	if merged.WaitFor != "updateContentCommand" {
		t.Fatalf("expected waitFor default updateContentCommand, got %q", merged.WaitFor)
	}
}

func TestMergeConfigurationAppliesShutdownActionDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config *DevContainerConfig
		want   string
	}{
		{
			name:   "single container",
			config: &DevContainerConfig{},
			want:   "stopContainer",
		},
		{
			name: "docker compose",
			config: &DevContainerConfig{
				ComposeContainer: ComposeContainer{
					DockerComposeFile: types.StrArray{"docker-compose.yml"},
					Service:           "app",
				},
			},
			want: "stopCompose",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			merged, err := MergeConfiguration(testCase.config, nil)
			if err != nil {
				t.Fatalf("MergeConfiguration() error = %v", err)
			}

			if merged.ShutdownAction != testCase.want {
				t.Fatalf("expected shutdownAction default %q, got %q", testCase.want, merged.ShutdownAction)
			}
		})
	}
}

func TestMergeConfigurationPreservesExplicitWaitForAndShutdownAction(t *testing.T) {
	tests := []struct {
		name    string
		config  *DevContainerConfig
		entries []*ImageMetadata
	}{
		{
			name: "parsed config",
			config: &DevContainerConfig{
				DevContainerConfigBase: DevContainerConfigBase{
					WaitFor:        "postCreateCommand",
					ShutdownAction: "none",
				},
			},
		},
		{
			name:   "image metadata",
			config: &DevContainerConfig{},
			entries: []*ImageMetadata{
				{
					DevContainerConfigBase: DevContainerConfigBase{
						WaitFor:        "postCreateCommand",
						ShutdownAction: "none",
					},
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			merged, err := MergeConfiguration(testCase.config, testCase.entries)
			if err != nil {
				t.Fatalf("MergeConfiguration() error = %v", err)
			}

			if merged.WaitFor != "postCreateCommand" {
				t.Fatalf("expected explicit waitFor to be preserved, got %q", merged.WaitFor)
			}
			if merged.ShutdownAction != "none" {
				t.Fatalf("expected explicit shutdownAction to be preserved, got %q", merged.ShutdownAction)
			}
		})
	}
}
