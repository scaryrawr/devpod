package cmd

import "testing"

func TestAddNPMRegistryEnv(t *testing.T) {
	t.Run("adds registry", func(t *testing.T) {
		got := addNPMRegistryEnv([]string{"EXAMPLE=value"}, "https://npm.example.com/")
		if len(got) != 2 || got[1] != "NPM_CONFIG_REGISTRY=https://npm.example.com/" {
			t.Fatalf("expected npm registry environment variable, got %#v", got)
		}
	})

	t.Run("preserves explicit registry", func(t *testing.T) {
		got := addNPMRegistryEnv(
			[]string{"npm_config_registry=https://explicit.example.com/"},
			"https://npm.example.com/",
		)
		if len(got) != 1 || got[0] != "npm_config_registry=https://explicit.example.com/" {
			t.Fatalf("expected explicit registry to be preserved, got %#v", got)
		}
	})
}

func TestContainsEnvKey(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		key  string
		want bool
	}{
		{
			name: "exact key",
			env:  []string{"NPM_CONFIG_REGISTRY=https://npm.example.com/"},
			key:  "NPM_CONFIG_REGISTRY",
			want: true,
		},
		{
			name: "npm lowercase key",
			env:  []string{"npm_config_registry=https://npm.example.com/"},
			key:  "NPM_CONFIG_REGISTRY",
			want: true,
		},
		{
			name: "different key",
			env:  []string{"NPM_CONFIG_USERCONFIG=/tmp/npmrc"},
			key:  "NPM_CONFIG_REGISTRY",
		},
		{
			name: "malformed value",
			env:  []string{"NPM_CONFIG_REGISTRY"},
			key:  "NPM_CONFIG_REGISTRY",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := containsEnvKey(test.env, test.key); got != test.want {
				t.Fatalf("expected %v, got %v", test.want, got)
			}
		})
	}
}
