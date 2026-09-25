package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type recordedCommand struct {
	Name  string
	Args  []string
	Input string
}

type fakeRunner struct {
	commands []recordedCommand
	stdout   map[string]string
	fail     map[string]error
}

func newFakeRunner() *fakeRunner {
	return &fakeRunner{stdout: map[string]string{}, fail: map[string]error{}}
}

func (f *fakeRunner) key(name string, args []string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	f.commands = append(f.commands, recordedCommand{Name: name, Args: args})
	key := f.key(name, args)

	return f.stdout[key], f.fail[key]
}

func (f *fakeRunner) RunWithInput(_ context.Context, input, name string, args ...string) error {
	f.commands = append(f.commands, recordedCommand{Name: name, Args: args, Input: input})

	return f.fail[f.key(name, args)]
}

func (f *fakeRunner) argv(index int) string {
	return f.key(f.commands[index].Name, f.commands[index].Args)
}

func testEnv() (*env, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer

	return &env{stdout: &stdout, stderr: &stderr}, &stdout, &stderr
}

func TestGetCurrentBranchTrims(t *testing.T) {
	runner := newFakeRunner()
	runner.stdout["git rev-parse --abbrev-ref HEAD"] = "APP-1-feature\n"

	branch, err := getCurrentBranch(context.Background(), runner)
	if err != nil {
		t.Fatalf("getCurrentBranch returned error: %v", err)
	}

	if want := "APP-1-feature"; branch != want {
		t.Errorf("branch = %q, want %q", branch, want)
	}
}

func TestGetCurrentBranchPropagatesFailure(t *testing.T) {
	runner := newFakeRunner()
	runner.fail["git rev-parse --abbrev-ref HEAD"] = errors.New("not a git repository")

	if _, err := getCurrentBranch(context.Background(), runner); err == nil {
		t.Fatal("getCurrentBranch succeeded, want error")
	}
}

func TestCreateBranch(t *testing.T) {
	runner := newFakeRunner()
	runner.stdout["git rev-parse --abbrev-ref HEAD"] = "main\n"
	e, stdout, _ := testEnv()

	if err := createBranch(context.Background(), runner, e, "APP-1-feature"); err != nil {
		t.Fatalf("createBranch returned error: %v", err)
	}

	if want := "git checkout -b APP-1-feature"; runner.argv(1) != want {
		t.Errorf("argv = %q, want %q", runner.argv(1), want)
	}
	if !strings.Contains(stdout.String(), "Switched to new branch: APP-1-feature") {
		t.Errorf("stdout = %q, want the switch message", stdout.String())
	}
}

func TestCreateBranchSkipsWhenAlreadyOnBranch(t *testing.T) {
	runner := newFakeRunner()
	runner.stdout["git rev-parse --abbrev-ref HEAD"] = "APP-1-feature\n"
	e, stdout, _ := testEnv()

	if err := createBranch(context.Background(), runner, e, "APP-1-feature"); err != nil {
		t.Fatalf("createBranch returned error: %v", err)
	}

	if len(runner.commands) != 1 {
		t.Errorf("ran %d commands, want only the branch lookup", len(runner.commands))
	}
	if !strings.Contains(stdout.String(), "Already on branch APP-1-feature") {
		t.Errorf("stdout = %q, want the already-on-branch message", stdout.String())
	}
}

func TestCreateBranchSwallowsCheckoutFailure(t *testing.T) {
	runner := newFakeRunner()
	runner.stdout["git rev-parse --abbrev-ref HEAD"] = "main\n"
	runner.fail["git checkout -b APP-1-feature"] = errors.New("branch exists")
	e, _, stderr := testEnv()

	if err := createBranch(context.Background(), runner, e, "APP-1-feature"); err != nil {
		t.Fatalf("createBranch returned error %v, want the failure reported but not returned", err)
	}

	if !strings.Contains(stderr.String(), "branch exists") {
		t.Errorf("stderr = %q, want the git failure reported", stderr.String())
	}
}

func TestExecRunnerAgainstRealGit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that shells out to git")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "--initial-branch", "APP-7-real"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		command := exec.Command("git", args...)
		command.Dir = dir
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, output)
		}
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(previous) })

	branch, err := getCurrentBranch(context.Background(), execRunner{})
	if err != nil {
		t.Fatalf("getCurrentBranch returned error: %v", err)
	}
	if want := "APP-7-real"; branch != want {
		t.Errorf("branch = %q, want %q", branch, want)
	}
}
