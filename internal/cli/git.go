package cli

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
	RunWithInput(ctx context.Context, input, name string, args ...string) error
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	command := exec.CommandContext(ctx, name, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", err
		}

		return "", errorWithOutput{err: err, output: message}
	}

	return stdout.String(), nil
}

func (execRunner) RunWithInput(ctx context.Context, input, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = strings.NewReader(input)

	return command.Run()
}

type errorWithOutput struct {
	err    error
	output string
}

func (e errorWithOutput) Error() string { return e.output }

func (e errorWithOutput) Unwrap() error { return e.err }

func getCurrentBranch(ctx context.Context, runner Runner) (string, error) {
	stdout, err := runner.Run(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout), nil
}

func createBranch(ctx context.Context, runner Runner, e *env, branchName string) error {
	currentBranch, err := getCurrentBranch(ctx, runner)
	if err != nil {
		return err
	}

	if currentBranch == branchName {
		e.info("Already on branch %s", branchName)

		return nil
	}

	if _, err := runner.Run(ctx, "git", "checkout", "-b", branchName); err != nil {
		e.failure(err)

		return nil
	}

	e.info("Switched to new branch: %s", branchName)

	return nil
}
