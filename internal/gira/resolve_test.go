package gira

import (
	"path/filepath"
	"testing"
)

func TestResolveProjectConfigUsesLongestMatchingPath(t *testing.T) {
	clearEnv(t)
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"defaults": {
			"gitlab": {"url": "https://gitlab.example.com", "targetBranch": "main"},
			"jira": {"issueType": "Task", "startStatus": "Doing"}
		},
		"projects": [
			{"path": "~/Projects/alpha", "gitlab": {"projectId": "111"}},
			{
				"path": "~/Projects/alpha/app",
				"gitlab": {"projectId": "222", "targetBranch": "develop"},
				"jira": {"enabled": false, "subtaskIssueType": "Sub-task", "reviewStatus": "Reviewing"}
			}
		]
	}`)
	cwd := filepath.Join(homeDir, "Projects", "alpha", "app", "src")

	resolved, err := ResolveProjectConfig(cwd, homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if got, want := resolved.Gitlab.URL, "https://gitlab.example.com"; got != want {
		t.Errorf("gitlab.url = %q, want %q", got, want)
	}
	if got, want := resolved.Gitlab.ProjectID, "222"; got != want {
		t.Errorf("gitlab.projectId = %q, want %q", got, want)
	}
	if got, want := resolved.Gitlab.TargetBranch, "develop"; got != want {
		t.Errorf("gitlab.targetBranch = %q, want %q", got, want)
	}
	if resolved.Jira.Enabled {
		t.Error("jira.enabled = true, want false")
	}
	if got, want := resolved.Jira.IssueType, "Task"; got != want {
		t.Errorf("jira.issueType = %q, want %q", got, want)
	}
	if got, want := resolved.Jira.SubtaskIssueType, "Sub-task"; got != want {
		t.Errorf("jira.subtaskIssueType = %q, want %q", got, want)
	}
	if got, want := resolved.Jira.StartStatus, "Doing"; got != want {
		t.Errorf("jira.startStatus = %q, want %q", got, want)
	}
	if got, want := resolved.Jira.ReviewStatus, "Reviewing"; got != want {
		t.Errorf("jira.reviewStatus = %q, want %q", got, want)
	}
	if resolved.MatchedProject == nil || resolved.MatchedProject.Path != "~/Projects/alpha/app" {
		t.Errorf("matched project = %+v, want ~/Projects/alpha/app", resolved.MatchedProject)
	}
}

func TestResolveProjectConfigKeepsDefaultTargetBranch(t *testing.T) {
	clearEnv(t)
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"defaults": {"gitlab": {"targetBranch": "main"}},
		"projects": [{"path": "~/Projects/app", "gitlab": {"projectId": "111"}}]
	}`)

	resolved, err := ResolveProjectConfig(filepath.Join(homeDir, "Projects", "app"), homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if got, want := resolved.Gitlab.ProjectID, "111"; got != want {
		t.Errorf("gitlab.projectId = %q, want %q", got, want)
	}
	if got, want := resolved.Gitlab.TargetBranch, "main"; got != want {
		t.Errorf("gitlab.targetBranch = %q, want %q", got, want)
	}
}

func TestResolveProjectConfigMatchesWorktreeBasePath(t *testing.T) {
	clearEnv(t)
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"projects": [{
			"path": "~/Projects/foo",
			"worktreeBasePath": "~/Projects/foo-worktrees",
			"gitlab": {"projectId": "111"},
			"jira": {"projectKey": "FOO"}
		}]
	}`)
	cwd := filepath.Join(homeDir, "Projects", "foo-worktrees", "branch-a")

	resolved, err := ResolveProjectConfig(cwd, homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if resolved.MatchedProject == nil {
		t.Fatal("no project matched the worktree base path")
	}
	if got, want := resolved.MatchedProject.Path, "~/Projects/foo"; got != want {
		t.Errorf("matched path = %q, want %q", got, want)
	}
	if got, want := resolved.MatchedProject.WorktreeBasePath, "~/Projects/foo-worktrees"; got != want {
		t.Errorf("matched worktreeBasePath = %q, want %q", got, want)
	}
	if got, want := resolved.Gitlab.ProjectID, "111"; got != want {
		t.Errorf("gitlab.projectId = %q, want %q", got, want)
	}
	if got, want := resolved.Jira.ProjectKey, "FOO"; got != want {
		t.Errorf("jira.projectKey = %q, want %q", got, want)
	}
}

func TestResolveProjectConfigKeepsEnvWhenConfigValuesBlank(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITLAB_URL", "https://gitlab.example.com")
	t.Setenv("GITLAB_PROJECT_ID", "999")
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"defaults": {"gitlab": {"url": ""}},
		"projects": [{"path": "~/Projects/app", "gitlab": {"projectId": ""}}]
	}`)

	resolved, err := ResolveProjectConfig(filepath.Join(homeDir, "Projects", "app"), homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if got, want := resolved.Gitlab.URL, "https://gitlab.example.com"; got != want {
		t.Errorf("gitlab.url = %q, want %q", got, want)
	}
	if got, want := resolved.Gitlab.ProjectID, "999"; got != want {
		t.Errorf("gitlab.projectId = %q, want %q", got, want)
	}
}

func TestResolveProjectConfigEqualLengthTieKeepsFileOrder(t *testing.T) {
	clearEnv(t)
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"projects": [
			{"path": "~/Projects/app", "gitlab": {"projectId": "first"}},
			{"path": "~/Projects/app", "gitlab": {"projectId": "second"}}
		]
	}`)

	resolved, err := ResolveProjectConfig(filepath.Join(homeDir, "Projects", "app"), homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if got, want := resolved.Gitlab.ProjectID, "first"; got != want {
		t.Errorf("gitlab.projectId = %q, want %q", got, want)
	}
}

func TestResolveProjectConfigOverridesWinOverEverything(t *testing.T) {
	clearEnv(t)
	t.Setenv("JIRA_PROJECT_KEY", "FROM-ENV")
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"defaults": {"jira": {"projectKey": "FROM-DEFAULTS"}},
		"projects": [{"path": "~/Projects/app", "jira": {"projectKey": "FROM-PROJECT"}}]
	}`)
	override := "FROM-FLAG"

	resolved, err := ResolveProjectConfig(
		filepath.Join(homeDir, "Projects", "app"),
		homeDir,
		Overrides{Jira: &JiraSection{ProjectKey: &override}},
	)
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if got, want := resolved.Jira.ProjectKey, "FROM-FLAG"; got != want {
		t.Errorf("jira.projectKey = %q, want %q", got, want)
	}
}

func TestResolveProjectConfigNoMatchFallsBackToDefaults(t *testing.T) {
	clearEnv(t)
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"defaults": {"gitlab": {"projectId": "default"}},
		"projects": [{"path": "~/Projects/other"}]
	}`)

	resolved, err := ResolveProjectConfig(filepath.Join(homeDir, "elsewhere"), homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if resolved.MatchedProject != nil {
		t.Errorf("matched project = %+v, want none", resolved.MatchedProject)
	}
	if got, want := resolved.Gitlab.ProjectID, "default"; got != want {
		t.Errorf("gitlab.projectId = %q, want %q", got, want)
	}
	if !resolved.Jira.Enabled {
		t.Error("jira.enabled = false, want true by default")
	}
	if got, want := resolved.Jira.StartStatus, DefaultStartStatus; got != want {
		t.Errorf("jira.startStatus = %q, want %q", got, want)
	}
	if got, want := resolved.Jira.ReviewStatus, DefaultReviewStatus; got != want {
		t.Errorf("jira.reviewStatus = %q, want %q", got, want)
	}
}

func TestResolveProjectConfigDoesNotMatchSiblingPrefix(t *testing.T) {
	clearEnv(t)
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{"projects": [{"path": "~/Projects/app", "gitlab": {"projectId": "111"}}]}`)

	resolved, err := ResolveProjectConfig(filepath.Join(homeDir, "Projects", "app-legacy"), homeDir, Overrides{})
	if err != nil {
		t.Fatalf("ResolveProjectConfig returned error: %v", err)
	}

	if resolved.MatchedProject != nil {
		t.Errorf("matched project = %+v, want none", resolved.MatchedProject)
	}
}

func clearEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"GITLAB_URL", "GITLAB_API_TOKEN", "GITLAB_PROJECT_ID", "GITLAB_USER_ID",
		"JIRA_URL", "JIRA_API_TOKEN", "JIRA_USER_EMAIL", "JIRA_USER_ID", "JIRA_PROJECT_KEY",
	} {
		t.Setenv(key, "")
	}
}
