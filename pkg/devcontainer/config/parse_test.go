package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveDevContainerJSON(t *testing.T) {
	type args struct {
		config *DevContainerConfig
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantJSON string
	}{
		{
			name: "test omit build field in devcontainer.json",
			args: args{
				config: &DevContainerConfig{
					ImageContainer: ImageContainer{
						Image: "test",
					},
				},
			},
			wantErr:  false,
			wantJSON: `{"image":"test"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := localTempDir(t)

			tt.args.config.Origin = filepath.Join(tmpDir, "devcontainer.json")

			if err := SaveDevContainerJSON(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("SaveDevContainerJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			contents, err := os.ReadFile(tt.args.config.Origin)
			if err != nil {
				t.Fatalf("Failed to read file contents: %v", err)
			}
			if string(contents) != tt.wantJSON {
				t.Errorf("Expected JSON = %v, got %v", tt.wantJSON, string(contents))
			}
		})
	}
}

func TestParseDevContainerJSONSpecFields(t *testing.T) {
	tmpDir := localTempDir(t)
	devContainerDir := filepath.Join(tmpDir, ".devcontainer")
	if err := os.MkdirAll(devContainerDir, 0755); err != nil {
		t.Fatalf("Failed to create devcontainer dir: %v", err)
	}

	configPath := filepath.Join(devContainerDir, "devcontainer.json")
	err := os.WriteFile(configPath, []byte(`{
		"$schema": "https://containers.dev/schemas/devContainer.schema.json",
		"name": "spec fields",
		"image": "ubuntu:latest",
		"portsAttributes": {
			"3000": {
				"label": "Application",
				"onAutoForward": "openBrowserOnce"
			}
		},
		"secrets": {
			"GITHUB_TOKEN": {
				"description": "GitHub token",
				"documentationUrl": "https://example.com/token"
			}
		},
		"hostRequirements": {
			"gpu": {
				"cores": 1000,
				"memory": "32gb"
			}
		}
	}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write devcontainer.json: %v", err)
	}

	parsed, err := ParseDevContainerJSON(tmpDir, "")
	if err != nil {
		t.Fatalf("ParseDevContainerJSON() error = %v", err)
	}
	if parsed.Schema != "https://containers.dev/schemas/devContainer.schema.json" {
		t.Fatalf("Expected schema to be preserved, got %q", parsed.Schema)
	}
	if parsed.PortsAttributes["3000"].OnAutoForward != "openBrowserOnce" {
		t.Fatalf("Expected portsAttributes to parse, got %#v", parsed.PortsAttributes)
	}
	if parsed.Secrets["GITHUB_TOKEN"].DocumentationURL != "https://example.com/token" {
		t.Fatalf("Expected secrets to parse, got %#v", parsed.Secrets)
	}
	if parsed.HostRequirements == nil || !parsed.HostRequirements.UsesGPU() || parsed.HostRequirements.GPU.Cores != 1000 || parsed.HostRequirements.GPU.Memory != "32gb" {
		t.Fatalf("Expected object GPU requirements to parse, got %#v", parsed.HostRequirements)
	}

	out, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("Marshal parsed config: %v", err)
	}
	var marshalled map[string]interface{}
	if err := json.Unmarshal(out, &marshalled); err != nil {
		t.Fatalf("Unmarshal marshalled config: %v", err)
	}
	if _, ok := marshalled["portsAttributes"]; !ok {
		t.Fatalf("Expected marshalled config to use portsAttributes, got %s", string(out))
	}
	if _, ok := marshalled["portAttributes"]; ok {
		t.Fatalf("Expected marshalled config not to use legacy portAttributes, got %s", string(out))
	}
}

func localTempDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp(".", "test-devcontainer-")
	if err != nil {
		t.Fatalf("Failed to create local temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})
	return tmpDir
}
