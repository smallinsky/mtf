package components

import (
	"reflect"
	"testing"
)

func Test_appendEnv(t *testing.T) {
	tests := []struct {
		name     string
		cmd      []string
		env      []string
		expected []string
	}{
		{
			name: "EmptyEnv",
			cmd: []string{
				"cmd",
				envPlaceholder,
				"run",
			},
			env: []string{},
			expected: []string{
				"cmd",
				"run",
			},
		},
		{
			name: "PlaceholrderInMiddle",
			cmd: []string{
				"cmd",
				envPlaceholder,
				"print",
			},
			env: []string{
				"path=home",
			},
			expected: []string{
				"cmd",
				"-e", "path=home",
				"print",
			},
		},
		{
			name: "PlaceholrderAtEnd",
			cmd: []string{
				"cmd",
				"help",
				envPlaceholder,
			},
			env: []string{
				"path=home",
			},
			expected: []string{
				"cmd",
				"help",
				"-e", "path=home",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendEnv(tt.cmd, tt.env); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got = %v, exp %v", got, tt.expected)
			}
		})
	}
}
