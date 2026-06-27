package setup

import (
	"slices"
	"testing"
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
