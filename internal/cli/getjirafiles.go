package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

func newGetJiraFilesCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "get-jira-files",
		Usage:        "Download attachments from a Jira issue",
		ArgsUsage:    "<issueKey>",
		OnUsageError: onUsageError,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output directory for attachments"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := requireArgs(c, 1); err != nil {
				return err
			}

			issueKey := c.Args().Get(0)

			jira, _, err := a.jira(gira.Overrides{})
			if err != nil {
				return err
			}

			attachments, err := jira.Attachments(ctx, issueKey)
			if err != nil {
				return err
			}

			if len(attachments) == 0 {
				a.e.info("No attachments found for issue %s", issueKey)

				return nil
			}

			baseDir := c.String("output")
			if baseDir == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				baseDir = filepath.Join(cwd, issueKey)
			}
			if err := os.MkdirAll(baseDir, 0o755); err != nil {
				return err
			}

			a.e.info("Downloading %d attachment(s) for %s...", len(attachments), issueKey)

			results := make([]attachmentResult, 0, len(attachments))
			for _, attachment := range attachments {
				outputPath := filepath.Join(baseDir, attachment.Filename)

				if err := jira.DownloadAttachment(ctx, attachment.Content, outputPath); err != nil {
					results = append(results, attachmentResult{
						Filename: attachment.Filename,
						Path:     outputPath,
						Error:    err.Error(),
					})
					a.e.failuref("✗ Failed to download: %s - %v", attachment.Filename, err)

					continue
				}

				results = append(results, attachmentResult{
					Filename: attachment.Filename,
					Path:     outputPath,
					Size:     attachment.Size,
					MimeType: attachment.MimeType,
					Success:  true,
				})
				a.e.info("✓ Downloaded: %s", attachment.Filename)
			}

			return printReport(a.e.stdout, attachmentReport{
				IssueKey:         issueKey,
				OutputDirectory:  baseDir,
				TotalAttachments: len(attachments),
				Results:          results,
			})
		},
	}
}
