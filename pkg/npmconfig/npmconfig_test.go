package npmconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserRegistry(t *testing.T) {
	tests := []struct {
		name    string
		npmrc   string
		want    string
		wantErr bool
	}{
		{
			name:  "custom registry",
			npmrc: "registry=https://npm.example.com/\n",
			want:  "https://npm.example.com/",
		},
		{
			name:  "only unscoped registry",
			npmrc: "@scope:registry=https://scoped.example.com/\nregistry=https://npm.example.com/\n//npm.example.com/:_authToken=secret\n",
			want:  "https://npm.example.com/",
		},
		{
			name:  "last registry wins",
			npmrc: "registry=https://first.example.com/\nregistry='https://second.example.com/'\n",
			want:  "https://second.example.com/",
		},
		{
			name:  "default registry ignored",
			npmrc: "registry=https://registry.npmjs.org\n",
		},
		{
			name:    "credentials rejected",
			npmrc:   "registry=https://user:password@npm.example.com/\n",
			wantErr: true,
		},
		{
			name:    "interpolation rejected",
			npmrc:   "registry=https://${NPM_TOKEN}@npm.example.com/\n",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			homeDir := t.TempDir()
			t.Setenv("HOME", homeDir)
			err := os.WriteFile(filepath.Join(homeDir, ".npmrc"), []byte(test.npmrc), 0600)
			if err != nil {
				t.Fatal(err)
			}

			got, err := UserRegistry()
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got registry %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}

func TestUserRegistryMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	registry, err := UserRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if registry != "" {
		t.Fatalf("expected no registry, got %q", registry)
	}
}
