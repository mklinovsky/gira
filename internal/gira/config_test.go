package gira

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const wantTemplate = `{
  "defaults": {
    "gitlab": {
      "url": "",
      "apiToken": "",
      "projectId": "",
      "userId": "",
      "targetBranch": ""
    },
    "jira": {
      "enabled": true,
      "url": "",
      "apiToken": "",
      "userEmail": "",
      "userId": "",
      "projectKey": "",
      "issueType": "",
      "subtaskIssueType": "",
      "startStatus": "",
      "reviewStatus": ""
    }
  },
  "projects": []
}
`

func TestLoadConfigMissingFile(t *testing.T) {
	homeDir := t.TempDir()

	_, err := LoadConfig(homeDir)

	if err == nil {
		t.Fatal("LoadConfig succeeded, want error")
	}
	if want := MissingConfigMessage(ConfigPath(homeDir)); err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestInitConfigWritesTemplate(t *testing.T) {
	homeDir := t.TempDir()

	configPath, err := InitConfig(homeDir, false)
	if err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading %s: %v", configPath, err)
	}
	if string(content) != wantTemplate {
		t.Errorf("template mismatch:\ngot:\n%s\nwant:\n%s", content, wantTemplate)
	}

	config, err := LoadConfig(homeDir)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if len(config.Projects) != 0 {
		t.Errorf("Projects = %v, want empty", config.Projects)
	}
	if config.Defaults == nil || config.Defaults.Jira == nil || config.Defaults.Jira.Enabled == nil || !*config.Defaults.Jira.Enabled {
		t.Fatal("defaults.jira.enabled did not survive the round trip as true")
	}
	if config.Defaults.Gitlab != nil {
		t.Errorf("defaults.gitlab = %+v, want dropped", config.Defaults.Gitlab)
	}
	for name, got := range map[string]*string{
		"jira.issueType":        config.Defaults.Jira.IssueType,
		"jira.subtaskIssueType": config.Defaults.Jira.SubtaskIssueType,
		"jira.startStatus":      config.Defaults.Jira.StartStatus,
		"jira.reviewStatus":     config.Defaults.Jira.ReviewStatus,
	} {
		if got != nil {
			t.Errorf("%s = %q, want unset", name, *got)
		}
	}
}

func TestInitConfigRequiresForceToOverwrite(t *testing.T) {
	homeDir := t.TempDir()

	configPath, err := InitConfig(homeDir, false)
	if err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"projects":[{"path":"/tmp/app"}]}`), 0o644); err != nil {
		t.Fatalf("overwriting config: %v", err)
	}

	_, err = InitConfig(homeDir, false)
	if err == nil {
		t.Fatal("second InitConfig succeeded, want error")
	}
	want := "Config file " + configPath + " already exists. Run gira init --force to overwrite it."
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}

	if _, err := InitConfig(homeDir, true); err != nil {
		t.Fatalf("InitConfig(force) returned error: %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading %s: %v", configPath, err)
	}
	if string(content) != wantTemplate {
		t.Errorf("forced init did not rewrite the template:\n%s", content)
	}
}

func TestLoadConfigKeepsEmptyValuesUnset(t *testing.T) {
	homeDir := t.TempDir()
	writeConfig(t, homeDir, `{
		"defaults": {
			"gitlab": {"url": "   ", "targetBranch": ""},
			"jira": {"enabled": false, "projectKey": "", "issueType": "", "subtaskIssueType": "   "}
		},
		"projects": []
	}`)

	config, err := LoadConfig(homeDir)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if config.Defaults.Gitlab != nil {
		t.Errorf("defaults.gitlab = %+v, want dropped: every value is blank", config.Defaults.Gitlab)
	}
	for name, got := range map[string]*string{
		"jira.projectKey":       config.Defaults.Jira.ProjectKey,
		"jira.issueType":        config.Defaults.Jira.IssueType,
		"jira.subtaskIssueType": config.Defaults.Jira.SubtaskIssueType,
	} {
		if got != nil {
			t.Errorf("%s = %q, want unset", name, *got)
		}
	}
	if config.Defaults.Jira.Enabled == nil || *config.Defaults.Jira.Enabled {
		t.Error("jira.enabled did not survive as false")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	homeDir := t.TempDir()
	configPath := writeConfig(t, homeDir, "{")

	_, err := LoadConfig(homeDir)

	if err == nil {
		t.Fatal("LoadConfig succeeded, want error")
	}
	if want := "Invalid JSON in " + configPath; !strings.HasPrefix(err.Error(), want) {
		t.Errorf("error = %q, want prefix %q", err.Error(), want)
	}
}

func TestLoadConfigIsLenient(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantProjects int
		check        func(*testing.T, Config)
	}{
		{
			name:    "non-string field is dropped, not fatal",
			content: `{"defaults":{"gitlab":{"projectId":12345,"url":"https://gitlab.example.com"}}}`,
			check: func(t *testing.T, c Config) {
				if c.Defaults.Gitlab.ProjectID != nil {
					t.Errorf("projectId = %q, want unset", *c.Defaults.Gitlab.ProjectID)
				}
				if c.Defaults.Gitlab.URL == nil || *c.Defaults.Gitlab.URL != "https://gitlab.example.com" {
					t.Error("url did not survive alongside the dropped field")
				}
			},
		},
		{name: "projects is not an array", content: `{"projects":{"path":"/tmp/app"}}`},
		{name: "project entry is a string", content: `{"projects":["/tmp/app"]}`},
		{name: "project without a path is skipped", content: `{"projects":[{"gitlab":{"projectId":"1"}}]}`},
		{name: "top level is an array", content: `[]`},
		{name: "top level is a string", content: `"nope"`},
		{
			name:         "usable project survives beside a broken one",
			content:      `{"projects":[{"gitlab":{}},{"path":"/tmp/app"}]}`,
			wantProjects: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			homeDir := t.TempDir()
			writeConfig(t, homeDir, tt.content)

			config, err := LoadConfig(homeDir)
			if err != nil {
				t.Fatalf("LoadConfig returned error: %v", err)
			}
			if len(config.Projects) != tt.wantProjects {
				t.Errorf("len(Projects) = %d, want %d", len(config.Projects), tt.wantProjects)
			}
			if tt.check != nil {
				tt.check(t, config)
			}
		})
	}
}

func writeConfig(t *testing.T, homeDir, content string) string {
	t.Helper()

	configPath := ConfigPath(homeDir)
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	return configPath
}
