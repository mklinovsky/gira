package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		secret string
		want   string
	}{
		{"", "(unset)"},
		{"a", "****"},
		{"glpat", "****"},
		{"1234567", "****"},
		{"12345678", "****5678"},
		{"glpat-AbCdEfGhIjKlMnOp", "****MnOp"},
	}

	for _, tt := range tests {
		t.Run(tt.secret, func(t *testing.T) {
			if got := maskSecret(tt.secret); got != tt.want {
				t.Errorf("maskSecret(%q) = %q, want %q", tt.secret, got, tt.want)
			}
		})
	}
}

func writeRawConfig(t *testing.T, homeDir, content string) string {
	t.Helper()

	configPath := filepath.Join(homeDir, ".gira", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	return configPath
}

func TestConfigCommandShowsWholeFile(t *testing.T) {
	homeDir := isolate(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	configPath := writeRawConfig(t, homeDir, `{
		"defaults": {
			"gitlab": {
				"url": "https://gitlab.example.com",
				"apiToken": "glpat-AbCdEfGhIjKlMnOp",
				"targetBranch": "main"
			},
			"jira": {
				"url": "https://example.atlassian.net",
				"apiToken": "jira-TokenValue1234",
				"userEmail": "me@example.com"
			}
		},
		"projects": [
			{
				"path": "`+cwd+`",
				"gitlab": {"projectId": "111", "apiToken": "project-TokenAbcWxyz"},
				"jira": {"projectKey": "HERE"}
			},
			{
				"path": "/somewhere/else",
				"worktreeBasePath": "/somewhere/else-worktrees",
				"jira": {"enabled": false, "projectKey": "THERE"}
			}
		]
	}`)
	h := newHarness()

	if code := h.run("config"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	out := h.stdout.String()

	for _, secret := range []string{"glpat-AbCdEfGhIjKlMnOp", "jira-TokenValue1234", "project-TokenAbcWxyz"} {
		if strings.Contains(out, secret) {
			t.Errorf("output leaks the token %q:\n%s", secret, out)
		}
	}

	for _, want := range []string{
		configPath,
		"****MnOp",
		"****1234",
		"****Wxyz",
		"https://gitlab.example.com",
		"me@example.com",
		"HERE",
		"/somewhere/else",
		"/somewhere/else-worktrees",
		"THERE",
		"false",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output does not contain %q:\n%s", want, out)
		}
	}
}

func TestConfigCommandMarksMatchingProject(t *testing.T) {
	homeDir := isolate(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	writeRawConfig(t, homeDir, `{"projects":[
		{"path":"/somewhere/else","jira":{"projectKey":"THERE"}},
		{"path":"`+cwd+`","jira":{"projectKey":"HERE"}}
	]}`)
	h := newHarness()

	if code := h.run("config"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	out := h.stdout.String()
	marked := ""
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, currentDirectoryMarker) {
			marked = line
		}
	}

	if marked == "" {
		t.Fatalf("no project is marked as matching the current directory:\n%s", out)
	}
	if !strings.Contains(marked, cwd) {
		t.Errorf("marked line = %q, want it to name %q", marked, cwd)
	}
}

func TestConfigCommandWithoutProjects(t *testing.T) {
	homeDir := isolate(t)
	stub := newAPIStub(t, map[string]string{})
	writeGiraConfig(t, homeDir, stub.server.URL)
	h := newHarness()

	if code := h.run("config"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	if !strings.Contains(h.stdout.String(), "(none)") {
		t.Errorf("output does not report the empty project list:\n%s", h.stdout.String())
	}
}

func TestConfigCommandShowsResolvedValues(t *testing.T) {
	homeDir := isolate(t)
	t.Setenv("JIRA_PROJECT_KEY", "FROM-ENV")
	writeRawConfig(t, homeDir, `{"defaults":{"jira":{"issueType":"Task"}},"projects":[]}`)
	h := newHarness()

	if code := h.run("config"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	out := h.stdout.String()
	resolved := out[strings.Index(out, "Resolved"):]
	if !strings.Contains(resolved, "FROM-ENV") {
		t.Errorf("resolved section does not show the environment override:\n%s", out)
	}
	if !strings.Contains(resolved, "(unset)") {
		t.Errorf("resolved section does not mark unset values:\n%s", out)
	}
}

func TestConfigCommandWithoutConfigFile(t *testing.T) {
	isolate(t)
	h := newHarness()

	if code := h.run("config"); code != exitRuntime {
		t.Errorf("exit code = %d, want %d", code, exitRuntime)
	}
	if !strings.Contains(h.stderr.String(), "Run gira init.") {
		t.Errorf("stderr = %q, want the missing config message", h.stderr.String())
	}
}
