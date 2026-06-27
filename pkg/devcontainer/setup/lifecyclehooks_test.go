package setup

import (
	"os/user"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/types"
	"gotest.tools/assert"
)

func TestGetLifecycleHookCommandArgs(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		expect []string
	}{
		{
			name:   "string command uses shell",
			input:  []string{"echo hello && echo goodbye"},
			expect: []string{"sh", "-c", "echo hello && echo goodbye"},
		},
		{
			name:   "array command executes directly",
			input:  []string{"echo", "hello && echo goodbye"},
			expect: []string{"echo", "hello && echo goodbye"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := getLifecycleHookCommandArgs(testCase.input)
			if !slices.Equal(got, testCase.expect) {
				t.Fatalf("expected %v, got %v", testCase.expect, got)
			}
		})
	}
}

func TestRunLifecycleHookRunsObjectCommandsInParallel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell command timing test requires sh")
	}

	currentUser, err := user.Current()
	assert.NilError(t, err)

	hook := types.LifecycleHook{
		"one": []string{"sh", "-c", "sleep 1"},
		"two": []string{"sh", "-c", "sleep 1"},
	}

	start := time.Now()
	err = runLifecycleHook(hook, currentUser.Username, currentUser.Username, ".", nil, "postStartCommand", log.Discard)
	assert.NilError(t, err)

	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Fatalf("expected object lifecycle entries to run in parallel, took %s", elapsed)
	}
}
