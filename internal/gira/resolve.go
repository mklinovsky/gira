package gira

import (
	"os"
	"strings"
)

type ResolvedGitlab struct {
	URL          string
	APIToken     string
	ProjectID    string
	UserID       string
	TargetBranch string
}

type ResolvedJira struct {
	Enabled          bool
	URL              string
	APIToken         string
	UserEmail        string
	UserID           string
	ProjectKey       string
	IssueType        string
	SubtaskIssueType string
	StartStatus      string
	ReviewStatus     string
}

const (
	DefaultStartStatus  = "In Progress"
	DefaultReviewStatus = "In Review"
)

type Overrides struct {
	Gitlab *GitlabSection
	Jira   *JiraSection
}

type Resolved struct {
	ConfigPath     string
	MatchedProject *ProjectConfig
	Gitlab         ResolvedGitlab
	Jira           ResolvedJira
}

func ResolveProjectConfig(cwd, homeDir string, overrides Overrides) (Resolved, error) {
	config, err := LoadConfig(homeDir)
	if err != nil {
		return Resolved{}, err
	}

	matched := findMatchingProject(config.Projects, cwd, homeDir)

	resolved := Resolved{
		ConfigPath:     ConfigPath(homeDir),
		MatchedProject: matched,
		Jira: ResolvedJira{
			Enabled:      true,
			StartStatus:  DefaultStartStatus,
			ReviewStatus: DefaultReviewStatus,
		},
	}

	var defaultGitlab, projectGitlab *GitlabSection
	var defaultJira, projectJira *JiraSection
	if config.Defaults != nil {
		defaultGitlab, defaultJira = config.Defaults.Gitlab, config.Defaults.Jira
	}
	if matched != nil {
		projectGitlab, projectJira = matched.Gitlab, matched.Jira
	}

	for _, section := range []*GitlabSection{gitlabEnvSection(), defaultGitlab, projectGitlab, overrides.Gitlab} {
		resolved.Gitlab.merge(section)
	}
	for _, section := range []*JiraSection{jiraEnvSection(), defaultJira, projectJira, overrides.Jira} {
		resolved.Jira.merge(section)
	}

	return resolved, nil
}

func (r *ResolvedGitlab) merge(section *GitlabSection) {
	if section == nil {
		return
	}

	setIfPresent(&r.URL, section.URL)
	setIfPresent(&r.APIToken, section.APIToken)
	setIfPresent(&r.ProjectID, section.ProjectID)
	setIfPresent(&r.UserID, section.UserID)
	setIfPresent(&r.TargetBranch, section.TargetBranch)
}

func (r *ResolvedJira) merge(section *JiraSection) {
	if section == nil {
		return
	}

	if section.Enabled != nil {
		r.Enabled = *section.Enabled
	}
	setIfPresent(&r.URL, section.URL)
	setIfPresent(&r.APIToken, section.APIToken)
	setIfPresent(&r.UserEmail, section.UserEmail)
	setIfPresent(&r.UserID, section.UserID)
	setIfPresent(&r.ProjectKey, section.ProjectKey)
	setIfPresent(&r.IssueType, section.IssueType)
	setIfPresent(&r.SubtaskIssueType, section.SubtaskIssueType)
	setIfPresent(&r.StartStatus, section.StartStatus)
	setIfPresent(&r.ReviewStatus, section.ReviewStatus)
}

func setIfPresent(target *string, value *string) {
	if value != nil {
		*target = *value
	}
}

func gitlabEnvSection() *GitlabSection {
	return &GitlabSection{
		URL:       envValue("GITLAB_URL"),
		APIToken:  envValue("GITLAB_API_TOKEN"),
		ProjectID: envValue("GITLAB_PROJECT_ID"),
		UserID:    envValue("GITLAB_USER_ID"),
	}
}

func jiraEnvSection() *JiraSection {
	enabled := true

	return &JiraSection{
		Enabled:    &enabled,
		URL:        envValue("JIRA_URL"),
		APIToken:   envValue("JIRA_API_TOKEN"),
		UserEmail:  envValue("JIRA_USER_EMAIL"),
		UserID:     envValue("JIRA_USER_ID"),
		ProjectKey: envValue("JIRA_PROJECT_KEY"),
	}
}

func envValue(key string) *string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	return &value
}
