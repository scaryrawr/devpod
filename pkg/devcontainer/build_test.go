package devcontainer

import (
	"runtime"
	"testing"
	"time"

	"github.com/loft-sh/devpod/pkg/devcontainer/config"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/types"
	"gotest.tools/assert"
)

func TestFallbackContainerContextAndDockerfile(t *testing.T) {
	testCases := []struct {
		name string

		localFolder  string
		remoteFolder string
		context      string
		dockerfile   string

		expectedContext    string
		expectedDockerfile string
	}{
		{
			name: "simple",

			localFolder:  "/my/local/folder",
			remoteFolder: "/workspaces/test",
			context:      "/my/local/folder/context",
			dockerfile:   "/my/local/folder/Dockerfile",

			expectedContext:    "/workspaces/test/context",
			expectedDockerfile: "/workspaces/test/Dockerfile",
		},
		{
			name: "windows",

			localFolder:  "C:/my/local/folder",
			remoteFolder: "/workspaces/test",
			context:      "C:/my/local/folder",
			dockerfile:   "C:/my/local/folder/Dockerfile",

			expectedContext:    "/workspaces/test",
			expectedDockerfile: "/workspaces/test/Dockerfile",
		},
	}

	for _, testCase := range testCases {
		outContext, outDockerfile := getContainerContextAndDockerfile(testCase.localFolder, testCase.remoteFolder, testCase.context, testCase.dockerfile)
		assert.Equal(t, outContext, testCase.expectedContext, testCase.name)
		assert.Equal(t, outDockerfile, testCase.expectedDockerfile, testCase.name)
	}
}

func TestRunInitializeCommandHookRunsObjectCommandsInParallel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell command timing test requires sh")
	}

	hook := types.LifecycleHook{
		"one": []string{"sh", "-c", "sleep 1"},
		"two": []string{"sh", "-c", "sleep 1"},
	}

	start := time.Now()
	err := runInitializeCommandHook(hook, []string{"sh", "-c"}, ".", nil, log.Discard)
	assert.NilError(t, err)

	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Fatalf("expected object initializeCommand entries to run in parallel, took %s", elapsed)
	}
}

func TestGetWorkspaceDockerComposeDefaultsWorkspaceFolderToRoot(t *testing.T) {
	workspaceMount, containerWorkspaceFolder := getWorkspace("/local/project", "project-id", &config.DevContainerConfig{
		ComposeContainer: config.ComposeContainer{
			DockerComposeFile: types.StrArray{"docker-compose.yml"},
			Service:           "app",
		},
	})

	parsedMount := config.ParseMount(workspaceMount)
	assert.Equal(t, containerWorkspaceFolder, "/")
	assert.Equal(t, parsedMount.Target, "/")
}

func TestGetWorkspaceDockerComposeUsesConfiguredWorkspaceFolder(t *testing.T) {
	workspaceMount, containerWorkspaceFolder := getWorkspace("/local/project", "project-id", &config.DevContainerConfig{
		DevContainerConfigBase: config.DevContainerConfigBase{
			WorkspaceFolder: "/workspace/app",
		},
		ComposeContainer: config.ComposeContainer{
			DockerComposeFile: types.StrArray{"docker-compose.yml"},
			Service:           "app",
		},
	})

	parsedMount := config.ParseMount(workspaceMount)
	assert.Equal(t, containerWorkspaceFolder, "/workspace/app")
	assert.Equal(t, parsedMount.Target, "/workspace/app")
}
