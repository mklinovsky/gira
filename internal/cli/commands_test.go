package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type apiStub struct {
	server   *httptest.Server
	requests []recordedRequest
}

type recordedRequest struct {
	Method string
	Path   string
	Body   string
}

func newAPIStub(t *testing.T, routes map[string]string) *apiStub {
	t.Helper()

	stub := &apiStub{}
	stub.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		stub.requests = append(stub.requests, recordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   string(body),
		})

		response, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			response, ok = routes[r.URL.Path]
		}
		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusTeapot)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(response))
	}))
	t.Cleanup(stub.server.Close)

	return stub
}

func (s *apiStub) bodyFor(t *testing.T, method, path string) map[string]any {
	t.Helper()

	for _, request := range s.requests {
		if request.Method == method && request.Path == path {
			var decoded map[string]any
			if err := json.Unmarshal([]byte(request.Body), &decoded); err != nil {
				t.Fatalf("decoding %s %s body %q: %v", method, path, request.Body, err)
			}

			return decoded
		}
	}

	t.Fatalf("no %s %s request was made", method, path)

	return nil
}

func isolate(t *testing.T) string {
	t.Helper()

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	for _, key := range []string{
		"GITLAB_URL", "GITLAB_API_TOKEN", "GITLAB_PROJECT_ID", "GITLAB_USER_ID",
		"JIRA_URL", "JIRA_API_TOKEN", "JIRA_USER_EMAIL", "JIRA_USER_ID", "JIRA_PROJECT_KEY",
	} {
		t.Setenv(key, "")
	}

	return homeDir
}

func writeGiraConfig(t *testing.T, homeDir, baseURL string) {
	t.Helper()

	configPath := filepath.Join(homeDir, ".gira", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}

	content := `{
		"defaults": {
			"gitlab": {
				"url": "` + baseURL + `",
				"apiToken": "glpat",
				"projectId": "42",
				"userId": "7",
				"targetBranch": "main"
			},
			"jira": {
				"enabled": true,
				"url": "` + baseURL + `",
				"apiToken": "token",
				"userEmail": "me@example.com",
				"userId": "user-1",
				"projectKey": "APP",
				"issueType": "Task"
			}
		},
		"projects": []
	}`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
}

type harness struct {
	app    *app
	runner *fakeRunner
	stdout *strings.Builder
	stderr *strings.Builder
}

func newHarness() *harness {
	runner := newFakeRunner()
	stdout := &strings.Builder{}
	stderr := &strings.Builder{}

	return &harness{
		app: &app{
			e:      &env{stdout: stdout, stderr: stderr},
			runner: runner,
			goos:   "darwin",
			client: http.DefaultClient,
		},
		runner: runner,
		stdout: stdout,
		stderr: stderr,
	}
}

func (h *harness) run(args ...string) int {
	return runWith(args, "0.5.0", h.app)
}

func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"update without custom field", []string{"update"}, "No custom field provided."},
		{"status without an argument", []string{"status"}, "gira status requires"},
		{"get-jira with too many arguments", []string{"get-jira", "APP-1", "APP-2"}, "gira get-jira requires"},
		{"malformed custom field", []string{"update", "-i", "APP-1", "--custom-field", "keyonly"}, "Invalid custom field format."},
		{"unknown flag", []string{"create", "Summary", "--nope"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolate(t)
			h := newHarness()

			if code := h.run(tt.args...); code != exitUsage {
				t.Errorf("exit code = %d, want %d (stderr %q)", code, exitUsage, h.stderr.String())
			}
			if tt.want != "" && !strings.Contains(h.stderr.String(), tt.want) {
				t.Errorf("stderr = %q, want it to contain %q", h.stderr.String(), tt.want)
			}
		})
	}
}

func TestMissingConfigIsRuntimeError(t *testing.T) {
	isolate(t)
	h := newHarness()

	if code := h.run("get-jira", "APP-1"); code != exitRuntime {
		t.Errorf("exit code = %d, want %d", code, exitRuntime)
	}
	if !strings.Contains(h.stderr.String(), "Run gira init.") {
		t.Errorf("stderr = %q, want the missing config message", h.stderr.String())
	}
	if h.stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", h.stdout.String())
	}
}

func TestInitCreatesConfig(t *testing.T) {
	homeDir := isolate(t)
	h := newHarness()

	if code := h.run("init"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".gira", "config.json")); err != nil {
		t.Fatalf("config was not created: %v", err)
	}
	if !strings.Contains(h.stdout.String(), "Created config at ") {
		t.Errorf("stdout = %q, want the created message", h.stdout.String())
	}
}

func TestCreateWithBranchAndBareStart(t *testing.T) {
	homeDir := isolate(t)
	stub := newAPIStub(t, map[string]string{
		"POST /rest/api/3/issue":                  `{"key":"APP-2"}`,
		"GET /rest/api/3/issue/APP-2/transitions": `{"transitions":[{"id":"21","name":"In Progress"}]}`,
		"POST /rest/api/3/issue/APP-2/transitions": `{}`,
	})
	writeGiraConfig(t, homeDir, stub.server.URL)
	h := newHarness()
	h.runner.stdout["git rev-parse --abbrev-ref HEAD"] = "main\n"

	if code := h.run("create", "Fix login flow", "-b", "-s"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	if want := "git checkout -b APP-2-fix-login-flow"; h.runner.argv(1) != want {
		t.Errorf("argv = %q, want %q", h.runner.argv(1), want)
	}
	if !strings.Contains(h.stdout.String(), "Changed status of issue APP-2 to In Progress") {
		t.Errorf("stdout = %q, want the default start status", h.stdout.String())
	}
}

func TestCreateBareStartUsesConfiguredStatus(t *testing.T) {
	homeDir := isolate(t)
	stub := newAPIStub(t, map[string]string{
		"POST /rest/api/3/issue":                   `{"key":"APP-4"}`,
		"GET /rest/api/3/issue/APP-4/transitions":  `{"transitions":[{"id":"22","name":"Doing"}]}`,
		"POST /rest/api/3/issue/APP-4/transitions": `{}`,
	})
	writeRawConfig(t, homeDir, `{"defaults":{"jira":{
		"url":"`+stub.server.URL+`","apiToken":"token","userEmail":"me@example.com",
		"userId":"user-1","projectKey":"APP","issueType":"Task","startStatus":"Doing"
	}},"projects":[]}`)
	h := newHarness()

	if code := h.run("create", "Configured status", "-sa"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	fields := stub.bodyFor(t, http.MethodPost, "/rest/api/3/issue")["fields"].(map[string]any)
	if _, assigned := fields["assignee"]; !assigned {
		t.Error("payload has no assignee, want -a to have been applied")
	}
	if !strings.Contains(h.stdout.String(), "Changed status of issue APP-4 to Doing") {
		t.Errorf("stdout = %q, want the configured start status", h.stdout.String())
	}
}

func TestCreateWithGroupedShortFlags(t *testing.T) {
	homeDir := isolate(t)
	stub := newAPIStub(t, map[string]string{
		"POST /rest/api/3/issue":                   `{"key":"APP-3"}`,
		"GET /rest/api/3/issue/APP-3/transitions":  `{"transitions":[{"id":"21","name":"In Progress"}]}`,
		"POST /rest/api/3/issue/APP-3/transitions": `{}`,
	})
	writeGiraConfig(t, homeDir, stub.server.URL)
	h := newHarness()
	h.runner.stdout["git rev-parse --abbrev-ref HEAD"] = "main\n"

	if code := h.run("create", "Grouped flags", "-sab"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	fields := stub.bodyFor(t, http.MethodPost, "/rest/api/3/issue")["fields"].(map[string]any)
	if _, assigned := fields["assignee"]; !assigned {
		t.Error("payload has no assignee, want -a to have been applied")
	}
	if want := "git checkout -b APP-3-grouped-flags"; h.runner.argv(1) != want {
		t.Errorf("argv = %q, want %q", h.runner.argv(1), want)
	}
	if !strings.Contains(h.stdout.String(), "Changed status of issue APP-3 to In Progress") {
		t.Errorf("stdout = %q, want the default start status", h.stdout.String())
	}
}

func TestMergeRequestFromBranchName(t *testing.T) {
	homeDir := isolate(t)
	stub := newAPIStub(t, map[string]string{
		"POST /api/v4/projects/42/merge_requests":   `{"web_url":"https://gitlab.example.com/mr/5"}`,
		"GET /rest/api/3/issue/APP-1/transitions":   `{"transitions":[{"id":"31","name":"In Review"}]}`,
		"POST /rest/api/3/issue/APP-1/transitions":  `{}`,
	})
	writeGiraConfig(t, homeDir, stub.server.URL)
	h := newHarness()
	h.runner.stdout["git rev-parse --abbrev-ref HEAD"] = "APP-1-my-feature\n"

	if code := h.run("mr", "-d"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	payload := stub.bodyFor(t, http.MethodPost, "/api/v4/projects/42/merge_requests")
	if got, want := payload["title"], "Draft: APP-1 My feature"; got != want {
		t.Errorf("title = %v, want %v", got, want)
	}
	if got, want := payload["target_branch"], "main"; got != want {
		t.Errorf("target_branch = %v, want %v", got, want)
	}
	if want := "pbcopy"; h.runner.argv(1) != want {
		t.Errorf("argv = %q, want %q", h.runner.argv(1), want)
	}
	if got := h.runner.commands[1].Input; got != "https://gitlab.example.com/mr/5\n" {
		t.Errorf("clipboard input = %q, want the url with a trailing newline", got)
	}
	if !strings.Contains(h.stdout.String(), "Changed status of issue APP-1 to In Review") {
		t.Errorf("stdout = %q, want the In Review status", h.stdout.String())
	}
}

func TestGetJiraPrintsServerJSONUnchanged(t *testing.T) {
	homeDir := isolate(t)
	stub := newAPIStub(t, map[string]string{
		"GET /rest/api/3/issue/APP-1": `{"zeta":1,"id":1234567890123456789,"fields":{"summary":"a < b"}}`,
	})
	writeGiraConfig(t, homeDir, stub.server.URL)
	h := newHarness()

	if code := h.run("get-jira", "APP-1"); code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr %q)", code, exitOK, h.stderr.String())
	}

	want := `{
  "zeta": 1,
  "id": 1234567890123456789,
  "fields": {
    "summary": "a < b"
  }
}
`
	if h.stdout.String() != want {
		t.Errorf("stdout =\n%s\nwant\n%s", h.stdout.String(), want)
	}
}
