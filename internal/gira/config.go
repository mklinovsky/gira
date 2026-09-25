package gira

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type GitlabSection struct {
	URL          *string
	APIToken     *string
	ProjectID    *string
	UserID       *string
	TargetBranch *string
}

type JiraSection struct {
	Enabled          *bool
	URL              *string
	APIToken         *string
	UserEmail        *string
	UserID           *string
	ProjectKey       *string
	IssueType        *string
	SubtaskIssueType *string
	StartStatus      *string
	ReviewStatus     *string
}

type Defaults struct {
	Gitlab *GitlabSection
	Jira   *JiraSection
}

type ProjectConfig struct {
	Path             string
	WorktreeBasePath string
	Gitlab           *GitlabSection
	Jira             *JiraSection
}

type Config struct {
	Defaults *Defaults
	Projects []ProjectConfig
}

type configTemplate struct {
	Defaults struct {
		Gitlab struct {
			URL          string `json:"url"`
			APIToken     string `json:"apiToken"`
			ProjectID    string `json:"projectId"`
			UserID       string `json:"userId"`
			TargetBranch string `json:"targetBranch"`
		} `json:"gitlab"`
		Jira struct {
			Enabled          bool   `json:"enabled"`
			URL              string `json:"url"`
			APIToken         string `json:"apiToken"`
			UserEmail        string `json:"userEmail"`
			UserID           string `json:"userId"`
			ProjectKey       string `json:"projectKey"`
			IssueType        string `json:"issueType"`
			SubtaskIssueType string `json:"subtaskIssueType"`
			StartStatus      string `json:"startStatus"`
			ReviewStatus     string `json:"reviewStatus"`
		} `json:"jira"`
	} `json:"defaults"`
	Projects []ProjectConfig `json:"projects"`
}

func HomeDir() (string, error) {
	for _, key := range []string{"HOME", "USERPROFILE"} {
		if value := os.Getenv(key); value != "" {
			return value, nil
		}
	}

	return "", errors.New("Could not determine home directory.")
}

func ConfigPath(homeDir string) string {
	return filepath.Join(homeDir, ".gira", "config.json")
}

func MissingConfigMessage(configPath string) string {
	return fmt.Sprintf("Config file %s not found. Run gira init.", configPath)
}

func LoadConfig(homeDir string) (Config, error) {
	configPath := ConfigPath(homeDir)

	content, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, errors.New(MissingConfigMessage(configPath))
		}

		return Config{}, err
	}

	if !json.Valid(content) {
		return Config{}, fmt.Errorf("Invalid JSON in %s: %s", configPath, invalidJSONReason(content))
	}

	return parseConfig(content), nil
}

func InitConfig(homeDir string, force bool) (string, error) {
	configPath := ConfigPath(homeDir)

	if _, err := os.Stat(configPath); err == nil {
		if !force {
			return "", fmt.Errorf("Config file %s already exists. Run gira init --force to overwrite it.", configPath)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return "", err
	}

	var template configTemplate
	template.Defaults.Jira.Enabled = true
	template.Projects = []ProjectConfig{}

	content, err := encodeJSON(template, "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		return "", err
	}

	return configPath, nil
}

func parseConfig(content []byte) Config {
	config := Config{Projects: []ProjectConfig{}}

	root := asRecord(content)
	if root == nil {
		return config
	}

	config.Defaults = parseDefaults(root["defaults"])

	var projects []json.RawMessage
	if err := json.Unmarshal(root["projects"], &projects); err != nil {
		return config
	}

	for _, raw := range projects {
		if project := parseProject(raw); project != nil {
			config.Projects = append(config.Projects, *project)
		}
	}

	return config
}

func parseDefaults(raw json.RawMessage) *Defaults {
	record := asRecord(raw)
	if record == nil {
		return nil
	}

	gitlab := parseGitlabSection(record["gitlab"])
	jira := parseJiraSection(record["jira"])
	if gitlab == nil && jira == nil {
		return nil
	}

	return &Defaults{Gitlab: gitlab, Jira: jira}
}

func parseProject(raw json.RawMessage) *ProjectConfig {
	record := asRecord(raw)
	if record == nil {
		return nil
	}

	path := normalizeString(record["path"])
	if path == nil {
		return nil
	}

	project := ProjectConfig{
		Path:   *path,
		Gitlab: parseGitlabSection(record["gitlab"]),
		Jira:   parseJiraSection(record["jira"]),
	}
	if worktreeBasePath := normalizeString(record["worktreeBasePath"]); worktreeBasePath != nil {
		project.WorktreeBasePath = *worktreeBasePath
	}

	return &project
}

func parseGitlabSection(raw json.RawMessage) *GitlabSection {
	record := asRecord(raw)
	if record == nil {
		return nil
	}

	section := GitlabSection{
		URL:          normalizeString(record["url"]),
		APIToken:     normalizeString(record["apiToken"]),
		ProjectID:    normalizeString(record["projectId"]),
		UserID:       normalizeString(record["userId"]),
		TargetBranch: normalizeString(record["targetBranch"]),
	}

	if section == (GitlabSection{}) {
		return nil
	}

	return &section
}

func parseJiraSection(raw json.RawMessage) *JiraSection {
	record := asRecord(raw)
	if record == nil {
		return nil
	}

	section := JiraSection{
		Enabled:          normalizeBool(record["enabled"]),
		URL:              normalizeString(record["url"]),
		APIToken:         normalizeString(record["apiToken"]),
		UserEmail:        normalizeString(record["userEmail"]),
		UserID:           normalizeString(record["userId"]),
		ProjectKey:       normalizeString(record["projectKey"]),
		IssueType:        normalizeString(record["issueType"]),
		SubtaskIssueType: normalizeString(record["subtaskIssueType"]),
		StartStatus:      normalizeString(record["startStatus"]),
		ReviewStatus:     normalizeString(record["reviewStatus"]),
	}

	if section == (JiraSection{}) {
		return nil
	}

	return &section
}

func asRecord(raw json.RawMessage) map[string]json.RawMessage {
	if len(raw) == 0 {
		return nil
	}

	var record map[string]json.RawMessage
	if err := json.Unmarshal(raw, &record); err != nil {
		return nil
	}

	return record
}

func normalizeString(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func normalizeBool(raw json.RawMessage) *bool {
	if len(raw) == 0 {
		return nil
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}

	return &value
}

func invalidJSONReason(content []byte) string {
	var probe any
	if err := json.Unmarshal(content, &probe); err != nil {
		return err.Error()
	}

	return "unexpected content"
}

func encodeJSON(value any, indent string) ([]byte, error) {
	var buffer strings.Builder

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indent)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}

	return []byte(buffer.String()), nil
}
