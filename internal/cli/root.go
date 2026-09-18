package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/urfave/cli/v3"
)

const (
	exitOK      = 0
	exitRuntime = 1
	exitUsage   = 2
)

func Run(args []string, version string) int {
	return run(args, version, os.Stdout, os.Stderr)
}

func run(args []string, version string, stdout, stderr io.Writer) int {
	e := &env{stdout: stdout, stderr: stderr}

	return runWith(args, version, &app{
		e:      e,
		runner: execRunner{},
		goos:   runtime.GOOS,
		client: http.DefaultClient,
	})
}

func runWith(args []string, version string, a *app) int {
	e := a.e

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintln(cmd.Root().Writer, cmd.Version)
	}

	normalized := append([]string{"gira"}, NormalizeOptionalValue(args)...)

	err := newRootCommand(version, a).Run(context.Background(), normalized)

	switch {
	case err == nil:
		return exitOK
	case isUsageError(err):
		e.failure(err)
		return exitUsage
	default:
		e.failure(err)
		return exitRuntime
	}
}

func newRootCommand(version string, a *app) *cli.Command {
	return &cli.Command{
		Name:                   "gira",
		Usage:                  "CLI tool for managing Gitlab and JIRA",
		Version:                version,
		UseShortOptionHandling: true,
		Writer:                 a.e.stdout,
		ErrWriter:              a.e.stderr,
		OnUsageError:           onUsageError,
		ExitErrHandler:         func(context.Context, *cli.Command, error) {},
		Commands: []*cli.Command{
			newInitCommand(a),
			newConfigCommand(a),
			newCreateCommand(a),
			newUpdateCommand(a),
			newStatusCommand(a),
			newMrCommand(a),
			newGetMrCommand(a),
			newGetJiraCommand(a),
			newGetJiraFilesCommand(a),
			newMergeCommand(a),
		},
	}
}

func onUsageError(_ context.Context, _ *cli.Command, err error, _ bool) error {
	return usageError{err}
}

func requireArgs(c *cli.Command, count int) error {
	if c.Args().Len() == count {
		return nil
	}

	if count == 0 {
		return usagef("gira %s takes no arguments", c.Name)
	}

	return usagef("gira %s requires %s", c.Name, c.ArgsUsage)
}

type usageError struct {
	err error
}

func (e usageError) Error() string { return e.err.Error() }

func (e usageError) Unwrap() error { return e.err }

func usagef(format string, args ...any) error {
	return usageError{fmt.Errorf(format, args...)}
}

func isUsageError(err error) bool {
	var target usageError

	return errors.As(err, &target)
}
