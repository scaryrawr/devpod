package build

import (
	"testing"

	"github.com/loft-sh/devpod/pkg/devcontainer/config"
	"github.com/loft-sh/devpod/pkg/provider"
	"github.com/loft-sh/devpod/pkg/types"
	"gotest.tools/assert"
)

func TestNewOptionsIncludesDevContainerCacheFrom(t *testing.T) {
	parsedConfig := &config.SubstitutedConfig{
		Config: &config.DevContainerConfig{
			DockerfileContainer: config.DockerfileContainer{
				Build: &config.ConfigBuildOptions{
					CacheFrom: types.StrArray{"cache-image:latest", "type=registry,ref=registry-cache:latest"},
				},
			},
		},
	}

	buildOptions, err := NewOptions("Dockerfile", "FROM alpine", parsedConfig, nil, "devcontainer:latest", provider.BuildOptions{}, "hash")

	assert.NilError(t, err)
	assert.DeepEqual(t, buildOptions.CacheFrom, []string{"cache-image:latest", "type=registry,ref=registry-cache:latest"})
}

func TestNewOptionsCombinesRegistryCacheWithDevContainerCacheFrom(t *testing.T) {
	parsedConfig := &config.SubstitutedConfig{
		Config: &config.DevContainerConfig{
			DockerfileContainer: config.DockerfileContainer{
				Build: &config.ConfigBuildOptions{
					CacheFrom: types.StrArray{"cache-image:latest"},
				},
			},
		},
	}

	buildOptions, err := NewOptions("Dockerfile", "FROM alpine", parsedConfig, nil, "devcontainer:latest", provider.BuildOptions{
		RegistryCache: "registry-cache:latest",
	}, "hash")

	assert.NilError(t, err)
	assert.DeepEqual(t, buildOptions.CacheFrom, []string{"type=registry,ref=registry-cache:latest", "cache-image:latest"})
}
