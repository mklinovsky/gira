package cli

import (
	"context"
	"errors"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newMrCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "mr",
		Usage:        "Create a merge request",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "target", Aliases: []string{"t"}, Usage: "Target branch"},
			&cli.StringFlag{Name: "title", Usage: "Merge request title"},
			&cli.StringFlag{Name: "labels", Aliases: []string{"l"}, Usage: "Comma-separated labels for the merge request"},
			&cli.BoolFlag{Name: "draft", Aliases: []string{"d"}, Usage: "Create a draft merge request"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 0); err != nil {
				return err
			}

			sourceBranch, err := getCurrentBranch(ctx, a.runner)
			if err != nil {
				return err
			}
			if sourceBranch == "" {
				return errors.New("No current branch found.")
			}

			jiraKey := jiraKeyFromBranchName(sourceBranch)
			jiraSummary := jiraSummaryFromBranchName(sourceBranch)

			title := c.String("title")
			if title == "" {
				if jiraKey != "" && jiraSummary != "" {
					title = jiraKey + " " + jiraSummary
				} else {
					title, err = createTitleFromBranchName(sourceBranch)
					if err != nil {
						return err
					}
				}
			}
			if c.Bool("draft") {
				title = "Draft: " + title
			}

			gitlab, resolved, err := a.gitlab(true)
			if err != nil {
				return err
			}

			targetBranch := c.String("target")
			if targetBranch == "" {
				targetBranch = resolved.Gitlab.TargetBranch
			}
			if targetBranch == "" {
				targetBranch = "master"
			}

			url, err := gitlab.CreateMergeRequest(ctx, gira.CreateMergeRequestParams{
				SourceBranch: sourceBranch,
				TargetBranch: targetBranch,
				Title:        title,
				Labels:       c.String("labels"),
			})
			if err != nil {
				return err
			}

			if copyToClipboard(ctx, a.runner, a.goos, url) {
				a.e.success("MR created, link copied to clipboard: %s", url)
			} else {
				a.e.success("MR created: %s", url)
				a.e.info("Could not copy link to clipboard")
			}

			if !resolved.Jira.Enabled || jiraKey == "" {
				return nil
			}

			jira, err := jiraClient(resolved, a.client)
			if err != nil {
				return err
			}

			const statusName = "In Review"
			if err := jira.ChangeIssueStatus(ctx, jiraKey, statusName); err != nil {
				return err
			}

			a.e.success("Changed status of issue %s to %s", jiraKey, statusName)

			return nil
		},
	}
}
