package cli

import (
	"context"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newStatusCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "status",
		Usage:        "Change the status of a JIRA issue",
		ArgsUsage:    "<status>",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "issue", Aliases: []string{"i"}, Usage: "Issue key, if not provided, will parse key from current branch name"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 1); err != nil {
				return err
			}

			statusName := c.Args().Get(0)

			issueKey, err := a.issueKey(ctx, c.String("issue"))
			if err != nil {
				return err
			}

			jira, _, err := a.jira(gira.Overrides{})
			if err != nil {
				return err
			}

			if err := jira.ChangeIssueStatus(ctx, issueKey, statusName); err != nil {
				return err
			}

			a.e.success("Changed status of issue %s to %s", issueKey, statusName)

			return nil
		},
	}
}
