package docker

import (
	"testing"

	"gotest.tools/assert"
)

func TestAppendInitArg(t *testing.T) {
	enabled := true
	disabled := false

	assert.DeepEqual(t, appendInitArg([]string{"run"}, &enabled), []string{"run", "--init"})
	assert.DeepEqual(t, appendInitArg([]string{"run"}, &disabled), []string{"run"})
	assert.DeepEqual(t, appendInitArg([]string{"run"}, nil), []string{"run"})
}
