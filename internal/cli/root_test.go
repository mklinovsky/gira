package cli

import (
	"bytes"
	"errors"
	"testing"
)

func TestRunVersionPrintsBareVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"--version"}, "1.2.3", &stdout, &stderr)

	if code != exitOK {
		t.Errorf("exit code = %d, want %d", code, exitOK)
	}
	if got, want := stdout.String(), "1.2.3\n"; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunHelpExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if code := run([]string{"--help"}, "1.2.3", &stdout, &stderr); code != exitOK {
		t.Errorf("exit code = %d, want %d", code, exitOK)
	}
	if stdout.Len() == 0 {
		t.Error("stdout is empty, want help text")
	}
}

func TestRunUnknownFlagIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	if code := run([]string{"--nope"}, "1.2.3", &stdout, &stderr); code != exitUsage {
		t.Errorf("exit code = %d, want %d", code, exitUsage)
	}
	if stderr.Len() == 0 {
		t.Error("stderr is empty, want the failure reported")
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty: failures belong on stderr", stdout.String())
	}
}

func TestIsUsageError(t *testing.T) {
	plain := errors.New("boom")

	if isUsageError(plain) {
		t.Error("isUsageError(plain) = true, want false")
	}
	if !isUsageError(usagef("bad %s", "input")) {
		t.Error("isUsageError(usagef(...)) = false, want true")
	}
	if !isUsageError(usageError{plain}) {
		t.Error("isUsageError(usageError{}) = false, want true")
	}
	if !errors.Is(usageError{plain}, plain) {
		t.Error("usageError does not unwrap to its cause")
	}
}
