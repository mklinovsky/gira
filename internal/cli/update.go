package cli

import (
	"context"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newUpdateCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "update",
		Usage:        "Update a JIRA issue",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "issue", Aliases: []string{"i"}, Usage: "Issue key, if not provided, will parse key from current branch name"},
			&cli.StringFlag{Name: "custom-field", Usage: "Custom field in key=value format"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 0); err != nil {
				return err
			}

			customFieldArg := c.String("custom-field")
			if customFieldArg == "" {
				return usagef("No custom field provided.")
			}

			customField, err := gira.ParseCustomField(customFieldArg)
			if err != nil {
				return usageError{err}
			}

			issueKey, err := a.issueKey(ctx, c.String("issue"))
			if err != nil {
				return err
			}

			jira, _, err := a.jira(gira.Overrides{})
			if err != nil {
				return err
			}

			if err := jira.UpdateIssue(ctx, issueKey, customField); err != nil {
				return err
			}

			a.e.success("Issue %s updated.", issueKey)

			return nil
		},
	}
}
