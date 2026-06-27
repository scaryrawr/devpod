package config

import (
	"encoding/json"
	"testing"
)

func TestDevContainerConfigParsesPortsAttributes(t *testing.T) {
	raw := []byte(`{
		"portsAttributes": {
			"3000": {
				"label": "Application",
				"onAutoForward": "openBrowser",
				"requireLocalPort": true,
				"elevateIfNeeded": true,
				"protocol": "https"
			}
		}
	}`)

	cfg := &DevContainerConfig{}
	if err := json.Unmarshal(raw, cfg); err != nil {
		t.Fatalf("unmarshal devcontainer: %v", err)
	}

	port, ok := cfg.PortsAttributes["3000"]
	if !ok {
		t.Fatalf("expected portsAttributes entry for 3000")
	}
	if port.Label != "Application" ||
		port.OnAutoForward != "openBrowser" ||
		!port.RequireLocalPort ||
		!port.ElevateIfNeeded ||
		port.Protocol != "https" {
		t.Fatalf("unexpected port attributes: %#v", port)
	}
}

func TestDevContainerConfigParsesPortsAttributesWithoutDroppingOtherFields(t *testing.T) {
	raw := []byte(`{
		"image": "ubuntu:latest",
		"portsAttributes": {
			"3000": {
				"label": "Application"
			}
		},
		"onCreateCommand": "echo created",
		"containerEnv": {
			"CONTAINER": "true"
		}
	}`)

	cfg := &DevContainerConfig{}
	if err := json.Unmarshal(raw, cfg); err != nil {
		t.Fatalf("unmarshal devcontainer: %v", err)
	}

	if cfg.Image != "ubuntu:latest" {
		t.Fatalf("expected image to parse, got %q", cfg.Image)
	}
	if cfg.OnCreateCommand[""][0] != "echo created" {
		t.Fatalf("expected onCreateCommand to parse, got %#v", cfg.OnCreateCommand)
	}
	if cfg.ContainerEnv["CONTAINER"] != "true" {
		t.Fatalf("expected containerEnv to parse, got %#v", cfg.ContainerEnv)
	}
}

func TestDevContainerConfigMarshalsPortsAttributes(t *testing.T) {
	cfg := &DevContainerConfig{
		DevContainerConfigBase: DevContainerConfigBase{
			PortsAttributes: map[string]PortAttribute{
				"3000": {Label: "Application"},
			},
		},
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal devcontainer: %v", err)
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal marshalled devcontainer: %v", err)
	}
	if _, ok := out["portsAttributes"]; !ok {
		t.Fatalf("expected JSON key portsAttributes in %s", raw)
	}
	if _, ok := out["portAttributes"]; ok {
		t.Fatalf("did not expect non-spec JSON key portAttributes in %s", raw)
	}
}

func TestParseMountPreservesValueContainingEquals(t *testing.T) {
	mount := ParseMount("type=bind,source=/workspace/name=with-equals,target=/workspace,readonly")

	if mount.Type != "bind" {
		t.Fatalf("unexpected mount type: %q", mount.Type)
	}
	if mount.Source != "/workspace/name=with-equals" {
		t.Fatalf("unexpected mount source: %q", mount.Source)
	}
	if mount.Target != "/workspace" {
		t.Fatalf("unexpected mount target: %q", mount.Target)
	}
	if len(mount.Other) != 1 || mount.Other[0] != "readonly" {
		t.Fatalf("unexpected mount options: %#v", mount.Other)
	}
}

func TestSubstituteLocalEnvDefaultPreservesColons(t *testing.T) {
	raw := &DevContainerConfig{
		DevContainerConfigBase: DevContainerConfigBase{
			RemoteEnv: map[string]string{
				"URL":      "${localEnv:URL:http://localhost:3000}",
				"OVERRIDE": "${localEnv:OVERRIDE:http://localhost:3000}",
			},
		},
	}
	out := &DevContainerConfig{}

	err := Substitute(&SubstitutionContext{
		Env: map[string]string{
			"OVERRIDE": "https://example.com:8443",
		},
	}, raw, out)
	if err != nil {
		t.Fatalf("substitute devcontainer: %v", err)
	}

	if out.RemoteEnv["URL"] != "http://localhost:3000" {
		t.Fatalf("expected default URL with colons to be preserved, got %q", out.RemoteEnv["URL"])
	}
	if out.RemoteEnv["OVERRIDE"] != "https://example.com:8443" {
		t.Fatalf("expected environment override to win, got %q", out.RemoteEnv["OVERRIDE"])
	}
}

func TestMergeConfigurationDefaultsRemoteUserToContainerUser(t *testing.T) {
	merged, err := MergeConfiguration(&DevContainerConfig{
		NonComposeBase: NonComposeBase{
			ContainerUser: "node",
		},
	}, nil)
	if err != nil {
		t.Fatalf("merge config: %v", err)
	}

	if merged.ContainerUser != "node" {
		t.Fatalf("expected containerUser to be preserved, got %q", merged.ContainerUser)
	}
	if merged.RemoteUser != "node" {
		t.Fatalf("expected remoteUser to default to containerUser, got %q", merged.RemoteUser)
	}
}

func TestMergeConfigurationDoesNotOverrideExplicitRemoteUser(t *testing.T) {
	merged, err := MergeConfiguration(&DevContainerConfig{
		DevContainerConfigBase: DevContainerConfigBase{
			RemoteUser: "vscode",
		},
		NonComposeBase: NonComposeBase{
			ContainerUser: "node",
		},
	}, nil)
	if err != nil {
		t.Fatalf("merge config: %v", err)
	}

	if merged.RemoteUser != "vscode" {
		t.Fatalf("expected explicit remoteUser to be preserved, got %q", merged.RemoteUser)
	}
}

func TestGetRemoteUserDefaultsToContainerUserBeforeImageLabel(t *testing.T) {
	user := GetRemoteUser(&Result{
		MergedConfig: &MergedDevContainerConfig{
			NonComposeBase: NonComposeBase{
				ContainerUser: "node",
			},
		},
		ContainerDetails: &ContainerDetails{
			Config: ContainerDetailsConfig{
				Labels: map[string]string{
					UserLabel: "root",
				},
			},
		},
	})

	if user != "node" {
		t.Fatalf("expected remote user to default to containerUser, got %q", user)
	}
}
