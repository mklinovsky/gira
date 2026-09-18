package cli

import (
	"context"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newGetJiraCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "get-jira",
		Usage:        "Get details of a Jira issue",
		ArgsUsage:    "<issueKey>",
		OnUsageError: onUsageError,
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 1); err != nil {
				return err
			}

			jira, _, err := a.jira(gira.Overrides{})
			if err != nil {
				return err
			}

			raw, err := jira.IssueRaw(ctx, c.Args().Get(0))
			if err != nil {
				return err
			}

			return printJSON(a.e.stdout, raw)
		},
	}
}
