package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/mklinovsky/gira/internal/gira"
)

type app struct {
	e      *env
	runner Runner
	goos   string
	client *http.Client
}

func (a *app) resolve(overrides gira.Overrides) (gira.Resolved, error) {
	homeDir, err := gira.HomeDir()
	if err != nil {
		return gira.Resolved{}, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return gira.Resolved{}, err
	}

	return gira.ResolveProjectConfig(cwd, homeDir, overrides)
}

func (a *app) jira(overrides gira.Overrides) (*gira.Jira, gira.Resolved, error) {
	resolved, err := a.resolve(overrides)
	if err != nil {
		return nil, gira.Resolved{}, err
	}

	client, err := jiraClient(resolved, a.client)
	if err != nil {
		return nil, resolved, err
	}

	return client, resolved, nil
}

func jiraClient(resolved gira.Resolved, httpClient *http.Client) (*gira.Jira, error) {
	if !resolved.Jira.Enabled {
		return nil, errors.New("Jira is disabled for current folder.")
	}

	for _, required := range []struct{ name, value string }{
		{"url", resolved.Jira.URL},
		{"apiToken", resolved.Jira.APIToken},
		{"userEmail", resolved.Jira.UserEmail},
		{"userId", resolved.Jira.UserID},
		{"projectKey", resolved.Jira.ProjectKey},
	} {
		if required.value == "" {
			return nil, fmt.Errorf("Jira setting %s is required for current folder.", required.name)
		}
	}

	return gira.NewJira(gira.JiraConfig{
		BaseURL:          resolved.Jira.URL,
		APIToken:         resolved.Jira.APIToken,
		UserEmail:        resolved.Jira.UserEmail,
		UserID:           resolved.Jira.UserID,
		ProjectKey:       resolved.Jira.ProjectKey,
		IssueType:        resolved.Jira.IssueType,
		SubtaskIssueType: resolved.Jira.SubtaskIssueType,
	}, httpClient), nil
}

func (a *app) gitlab(requireUserID bool) (*gira.Gitlab, gira.Resolved, error) {
	resolved, err := a.resolve(gira.Overrides{})
	if err != nil {
		return nil, gira.Resolved{}, err
	}

	for _, required := range []struct{ name, value string }{
		{"url", resolved.Gitlab.URL},
		{"apiToken", resolved.Gitlab.APIToken},
		{"projectId", resolved.Gitlab.ProjectID},
	} {
		if required.value == "" {
			return nil, resolved, fmt.Errorf("GitLab setting %s is required for current folder.", required.name)
		}
	}

	if requireUserID && resolved.Gitlab.UserID == "" {
		return nil, resolved, errors.New("GitLab setting userId is required for current folder.")
	}

	return gira.NewGitlab(gira.GitlabConfig{
		BaseURL:   resolved.Gitlab.URL,
		APIToken:  resolved.Gitlab.APIToken,
		ProjectID: resolved.Gitlab.ProjectID,
		UserID:    resolved.Gitlab.UserID,
	}, a.client), resolved, nil
}

func (a *app) issueKey(ctx context.Context, flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	branch, err := getCurrentBranch(ctx, a.runner)
	if err != nil {
		return "", err
	}
	if branch == "" {
		return "", errors.New("No current branch found.")
	}

	issueKey := jiraKeyFromBranchName(branch)
	if issueKey == "" {
		return "", errors.New("No issue key found.")
	}

	return issueKey, nil
}
