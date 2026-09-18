package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

func newMergeCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "merge",
		Usage:        "Merge the current merge request",
		ArgsUsage:    "<mergeRequestId>",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "close", Aliases: []string{"c"}, Usage: "Close the JIRA issue after merging"},
			&cli.BoolFlag{Name: "delete", Aliases: []string{"d"}, Usage: "Delete the source branch after merging"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 1); err != nil {
				return err
			}

			mergeRequestID := c.Args().Get(0)

			gitlab, resolved, err := a.gitlab(false)
			if err != nil {
				return err
			}

			title, err := gitlab.MergeRequestTitle(ctx, mergeRequestID)
			if err != nil {
				return err
			}

			if err := gitlab.Merge(ctx, mergeRequestID, c.Bool("delete")); err != nil {
				return err
			}

			a.e.success("Merge request %s merged successfully.", mergeRequestID)

			jiraKey := jiraKeyFromBranchName(title)
			if !resolved.Jira.Enabled || jiraKey == "" || !c.Bool("close") {
				return nil
			}

			jira, err := jiraClient(resolved, a.client)
			if err != nil {
				return err
			}

			const statusName = "Done"
			if err := jira.ChangeIssueStatus(ctx, jiraKey, statusName); err != nil {
				return err
			}

			a.e.success("Changed status of issue %s to %s", jiraKey, statusName)

			return nil
		},
	}
}
