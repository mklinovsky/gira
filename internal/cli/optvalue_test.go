package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNormalizeOptionalValue(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "value follows the long flag",
			args: []string{"create", "sum", "--start", "In Progress"},
			want: []string{"create", "sum", "--start=In Progress"},
		},
		{
			name: "value follows the short flag",
			args: []string{"create", "sum", "-s", "In Progress"},
			want: []string{"create", "sum", "--start=In Progress"},
		},
		{
			name: "flag before the positional still takes the value",
			args: []string{"create", "-s", "In Progress", "sum"},
			want: []string{"create", "--start=In Progress", "sum"},
		},
		{
			name: "bare flag gets the default status",
			args: []string{"create", "sum", "-s"},
			want: []string{"create", "sum", "--start=In Progress"},
		},
		{
			name: "bare flag followed by another flag",
			args: []string{"create", "sum", "-s", "-a"},
			want: []string{"create", "sum", "--start=In Progress", "-a"},
		},
		{
			name: "grouped booleans are left alone",
			args: []string{"create", "sum", "-ba"},
			want: []string{"create", "sum", "-ba"},
		},
		{
			name: "group ending in s is split",
			args: []string{"create", "sum", "-bs", "Ready"},
			want: []string{"create", "sum", "-b", "--start=Ready"},
		},
		{
			name: "group ending in s with no value",
			args: []string{"create", "sum", "-bs"},
			want: []string{"create", "sum", "-b", "--start=In Progress"},
		},
		{
			name: "attached values pass through",
			args: []string{"create", "sum", "--start=Ready", "-s=Other", "-sThird"},
			want: []string{"create", "sum", "--start=Ready", "-s=Other", "-sThird"},
		},
		{
			name: "nothing after a double dash is rewritten",
			args: []string{"create", "--", "-s", "literal"},
			want: []string{"create", "--", "-s", "literal"},
		},
		{
			name: "other commands are untouched",
			args: []string{"status", "-s", "In Review"},
			want: []string{"status", "-s", "In Review"},
		},
		{
			name: "empty args",
			args: []string{},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, NormalizeOptionalValue(tt.args)); diff != "" {
				t.Errorf("NormalizeOptionalValue mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
