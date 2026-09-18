package cli

import (
	"context"
	"errors"
	"testing"
)

func TestCopyToClipboardPerPlatform(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"darwin", "pbcopy"},
		{"linux", "xclip -selection clipboard"},
		{"windows", "clip"},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			runner := newFakeRunner()

			if !copyToClipboard(context.Background(), runner, tt.goos, "https://gitlab.example.com/mr/1") {
				t.Fatal("copyToClipboard = false, want true")
			}
			if runner.argv(0) != tt.want {
				t.Errorf("argv = %q, want %q", runner.argv(0), tt.want)
			}
			if want := "https://gitlab.example.com/mr/1\n"; runner.commands[0].Input != want {
				t.Errorf("input = %q, want %q", runner.commands[0].Input, want)
			}
		})
	}
}

func TestCopyToClipboardFallsBackToXsel(t *testing.T) {
	runner := newFakeRunner()
	runner.fail["xclip -selection clipboard"] = errors.New("xclip not found")

	if !copyToClipboard(context.Background(), runner, "linux", "link") {
		t.Fatal("copyToClipboard = false, want true after the xsel fallback")
	}

	if want := "xsel --clipboard --input"; runner.argv(1) != want {
		t.Errorf("argv = %q, want %q", runner.argv(1), want)
	}
}

func TestCopyToClipboardReportsFailure(t *testing.T) {
	runner := newFakeRunner()
	runner.fail["pbcopy"] = errors.New("no pbcopy")

	if copyToClipboard(context.Background(), runner, "darwin", "link") {
		t.Error("copyToClipboard = true, want false")
	}
	if copyToClipboard(context.Background(), runner, "plan9", "link") {
		t.Error("copyToClipboard on an unsupported platform = true, want false")
	}
}
