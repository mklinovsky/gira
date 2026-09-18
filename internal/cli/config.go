package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mklinovsky/gira/internal/gira"
	"github.com/urfave/cli/v3"
)

const currentDirectoryMarker = "<- current directory"

func newConfigCommand(a *app) *cli.Command {
	return &cli.Command{
		Name:         "config",
		Usage:        "Show the config file with tokens masked",
		OnUsageError: onUsageError,
		Action: func(_ context.Context, c *cli.Command) error {
			if err := requireArgs(c, 0); err != nil {
				return err
			}

			homeDir, err := gira.HomeDir()
			if err != nil {
				return err
			}

			config, err := gira.LoadConfig(homeDir)
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			resolved, err := gira.ResolveProjectConfig(cwd, homeDir, gira.Overrides{})
			if err != nil {
				return err
			}

			out := a.e.stdout
			fmt.Fprintf(out, "Config file: %s\n", gira.ConfigPath(homeDir))

			fmt.Fprintf(out, "\nDefaults\n")
			if config.Defaults == nil {
				fmt.Fprintf(out, "  (none)\n")
			} else {
				printLayeredSections(out, "  ", config.Defaults.Gitlab, config.Defaults.Jira)
			}

			fmt.Fprintf(out, "\nProjects\n")
			if len(config.Projects) == 0 {
				fmt.Fprintf(out, "  (none)\n")
			}
			marked := false
			for index, project := range config.Projects {
				marker := ""
				if !marked && resolved.MatchedProject != nil && sameProject(*resolved.MatchedProject, project) {
					marker = "  " + currentDirectoryMarker
					marked = true
				}

				fmt.Fprintf(out, "  [%d] %s%s\n", index+1, project.Path, marker)
				if project.WorktreeBasePath != "" {
					fmt.Fprintf(out, "      worktreeBasePath: %s\n", project.WorktreeBasePath)
				}
				printLayeredSections(out, "      ", project.Gitlab, project.Jira)
			}

			fmt.Fprintf(out, "\nResolved for %s\n", cwd)
			printRows(out, "  ", "GitLab", [][2]string{
				{"url", resolved.Gitlab.URL},
				{"apiToken", maskSecret(resolved.Gitlab.APIToken)},
				{"projectId", resolved.Gitlab.ProjectID},
				{"userId", resolved.Gitlab.UserID},
				{"targetBranch", resolved.Gitlab.TargetBranch},
			})
			printRows(out, "  ", "Jira", [][2]string{
				{"enabled", strconv.FormatBool(resolved.Jira.Enabled)},
				{"url", resolved.Jira.URL},
				{"apiToken", maskSecret(resolved.Jira.APIToken)},
				{"userEmail", resolved.Jira.UserEmail},
				{"userId", resolved.Jira.UserID},
				{"projectKey", resolved.Jira.ProjectKey},
				{"issueType", resolved.Jira.IssueType},
				{"subtaskIssueType", resolved.Jira.SubtaskIssueType},
			})

			return nil
		},
	}
}

func sameProject(left, right gira.ProjectConfig) bool {
	return left.Path == right.Path && left.WorktreeBasePath == right.WorktreeBasePath
}

func printLayeredSections(out io.Writer, indent string, gitlab *gira.GitlabSection, jira *gira.JiraSection) {
	printRows(out, indent, "GitLab", gitlabRows(gitlab))
	printRows(out, indent, "Jira", jiraRows(jira))
}

func gitlabRows(section *gira.GitlabSection) [][2]string {
	if section == nil {
		return nil
	}

	rows := [][2]string{}
	rows = appendRow(rows, "url", section.URL, false)
	rows = appendRow(rows, "apiToken", section.APIToken, true)
	rows = appendRow(rows, "projectId", section.ProjectID, false)
	rows = appendRow(rows, "userId", section.UserID, false)
	rows = appendRow(rows, "targetBranch", section.TargetBranch, false)

	return rows
}

func jiraRows(section *gira.JiraSection) [][2]string {
	if section == nil {
		return nil
	}

	rows := [][2]string{}
	if section.Enabled != nil {
		rows = append(rows, [2]string{"enabled", strconv.FormatBool(*section.Enabled)})
	}
	rows = appendRow(rows, "url", section.URL, false)
	rows = appendRow(rows, "apiToken", section.APIToken, true)
	rows = appendRow(rows, "userEmail", section.UserEmail, false)
	rows = appendRow(rows, "userId", section.UserID, false)
	rows = appendRow(rows, "projectKey", section.ProjectKey, false)
	rows = appendRow(rows, "issueType", section.IssueType, false)
	rows = appendRow(rows, "subtaskIssueType", section.SubtaskIssueType, false)

	return rows
}

func appendRow(rows [][2]string, name string, value *string, secret bool) [][2]string {
	if value == nil {
		return rows
	}

	if secret {
		return append(rows, [2]string{name, maskSecret(*value)})
	}

	return append(rows, [2]string{name, *value})
}

func printRows(out io.Writer, indent, title string, rows [][2]string) {
	if len(rows) == 0 {
		return
	}

	width := 0
	for _, row := range rows {
		if len(row[0]) > width {
			width = len(row[0])
		}
	}

	fmt.Fprintf(out, "%s%s\n", indent, title)
	for _, row := range rows {
		value := row[1]
		if strings.TrimSpace(value) == "" {
			value = "(unset)"
		}
		fmt.Fprintf(out, "%s  %-*s  %s\n", indent, width, row[0], value)
	}
}

func maskSecret(secret string) string {
	if secret == "" {
		return "(unset)"
	}

	if len(secret) < 8 {
		return "****"
	}

	return "****" + secret[len(secret)-4:]
}
