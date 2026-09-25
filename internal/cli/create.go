package cli

import (
	"context"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newCreateCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "create",
		Usage:        "Create a new JIRA issue",
		ArgsUsage:    "<summary>",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "parent", Aliases: []string{"p"}, Usage: "Parent issue key"},
			&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Usage: "Issue type"},
			&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Usage: "Project key (overrides JIRA_PROJECT_KEY env variable)"},
			&cli.BoolFlag{Name: "branch", Aliases: []string{"b"}, Usage: "Create git branch"},
			&cli.BoolFlag{Name: "assign", Aliases: []string{"a"}, Usage: "Assign to me"},
			&cli.BoolFlag{Name: "start", Aliases: []string{"s"}, Usage: "Start progress (moves issue to jira.startStatus)"},
			&cli.StringFlag{Name: "custom-field", Usage: "Custom field in key=value format"},
			&cli.StringFlag{Name: "description", Aliases: []string{"d"}, Usage: "Issue description"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 1); err != nil {
				return err
			}

			summary := c.Args().Get(0)
			createBranchFlag := c.Bool("branch")

			customField, err := gira.ParseCustomField(c.String("custom-field"))
			if err != nil {
				return usageError{err}
			}

			overrides := gira.Overrides{}
			if projectKey := c.String("key"); projectKey != "" {
				overrides.Jira = &gira.JiraSection{ProjectKey: &projectKey}
			}

			jira, resolved, err := a.jira(overrides)
			if err != nil {
				return err
			}

			created, err := jira.CreateIssue(ctx, gira.CreateIssueParams{
				Summary:     summary,
				IssueType:   c.String("type"),
				ParentKey:   c.String("parent"),
				Description: c.String("description"),
				AssignToMe:  c.Bool("assign"),
				CustomField: customField,
			})
			if err != nil {
				return err
			}

			a.e.success("Issue created: %s", created.URL)

			if createBranchFlag {
				if err := createBranch(ctx, a.runner, a.e, createBranchName(created.Key, summary)); err != nil {
					return err
				}
			}

			if !c.Bool("start") {
				return nil
			}

			statusName := resolved.Jira.StartStatus

			if err := jira.ChangeIssueStatus(ctx, created.Key, statusName); err != nil {
				return err
			}

			a.e.success("Changed status of issue %s to %s", created.Key, statusName)

			return nil
		},
	}
}
