package gira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type recordedRequest struct {
	Method string
	Path   string
	Auth   string
	Body   string
}

type jiraStub struct {
	server   *httptest.Server
	requests []recordedRequest
}

func newJiraStub(t *testing.T, routes map[string]func(w http.ResponseWriter)) *jiraStub {
	t.Helper()

	stub := &jiraStub{}
	stub.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		stub.requests = append(stub.requests, recordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Auth:   r.Header.Get("Authorization"),
			Body:   string(body),
		})

		handler, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			handler, ok = routes[r.URL.Path]
		}
		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		handler(w)
	}))
	t.Cleanup(stub.server.Close)

	return stub
}

func (s *jiraStub) jira(config JiraConfig) *Jira {
	config.BaseURL = s.server.URL

	return NewJira(config, s.server.Client())
}

func jsonBody(t *testing.T, raw string) map[string]any {
	t.Helper()

	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatalf("decoding request body %q: %v", raw, err)
	}

	return decoded
}

func createdFields(t *testing.T, raw string) map[string]any {
	t.Helper()

	fields, ok := jsonBody(t, raw)["fields"].(map[string]any)
	if !ok {
		t.Fatalf("request body has no fields object: %s", raw)
	}

	return fields
}

func writeJSON(status int, body string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}
}

func baseJiraConfig() JiraConfig {
	return JiraConfig{
		APIToken:         "token",
		UserEmail:        "me@example.com",
		UserID:           "user-1",
		ProjectKey:       "APP",
		IssueType:        "Task",
		SubtaskIssueType: "Sub-task",
	}
}

func TestCreateIssueUsesSubtaskTypeForStandardParent(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1": writeJSON(200, `{"fields":{"issuetype":{"subtask":false,"hierarchyLevel":0}}}`),
		"/rest/api/3/issue":       writeJSON(201, `{"key":"APP-2"}`),
	})

	created, err := stub.jira(baseJiraConfig()).CreateIssue(context.Background(), CreateIssueParams{
		Summary:   "Child issue",
		ParentKey: "APP-1",
	})
	if err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}

	if created.Key != "APP-2" {
		t.Errorf("key = %q, want %q", created.Key, "APP-2")
	}
	if len(stub.requests) != 2 {
		t.Fatalf("made %d requests, want 2", len(stub.requests))
	}
	if got, want := stub.requests[0].Path, "/rest/api/3/issue/APP-1"; got != want {
		t.Errorf("first request path = %q, want %q", got, want)
	}

	fields := createdFields(t, stub.requests[1].Body)
	if got := fields["issuetype"].(map[string]any)["name"]; got != "Sub-task" {
		t.Errorf("issuetype.name = %v, want Sub-task", got)
	}
	if got := fields["parent"].(map[string]any)["key"]; got != "APP-1" {
		t.Errorf("parent.key = %v, want APP-1", got)
	}
}

func TestCreateIssueUsesStandardTypeForEpicParent(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1": writeJSON(200, `{"fields":{"issuetype":{"subtask":false,"hierarchyLevel":1}}}`),
		"/rest/api/3/issue":       writeJSON(201, `{"key":"APP-2"}`),
	})
	config := baseJiraConfig()
	config.IssueType = "Engineering Task"

	if _, err := stub.jira(config).CreateIssue(context.Background(), CreateIssueParams{
		Summary:   "Child issue",
		ParentKey: "APP-1",
	}); err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}

	fields := createdFields(t, stub.requests[1].Body)
	if got := fields["issuetype"].(map[string]any)["name"]; got != "Engineering Task" {
		t.Errorf("issuetype.name = %v, want Engineering Task", got)
	}
}

func TestCreateIssueFailsWithoutSubtaskType(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1": writeJSON(200, `{"fields":{"issuetype":{"subtask":false,"hierarchyLevel":0}}}`),
	})
	config := baseJiraConfig()
	config.SubtaskIssueType = ""

	_, err := stub.jira(config).CreateIssue(context.Background(), CreateIssueParams{
		Summary:   "Child issue",
		ParentKey: "APP-1",
	})

	if err == nil {
		t.Fatal("CreateIssue succeeded, want error")
	}
	want := "Jira setting subtaskIssueType is required when parent issue requires subtask creation."
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestCreateIssueFailsWhenParentIsSubtask(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1": writeJSON(200, `{"fields":{"issuetype":{"subtask":true,"hierarchyLevel":-1}}}`),
	})

	_, err := stub.jira(baseJiraConfig()).CreateIssue(context.Background(), CreateIssueParams{
		Summary:   "Child issue",
		ParentKey: "APP-1",
	})

	if err == nil {
		t.Fatal("CreateIssue succeeded, want error")
	}
	if want := "Cannot create child issue under subtask APP-1."; err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestCreateIssueSendsBasicAuthAndUrl(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue": writeJSON(201, `{"key":"APP-9"}`),
	})

	created, err := stub.jira(baseJiraConfig()).CreateIssue(context.Background(), CreateIssueParams{Summary: "Standalone"})
	if err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}

	if want := stub.server.URL + "/browse/APP-9"; created.URL != want {
		t.Errorf("url = %q, want %q", created.URL, want)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("me@example.com:token"))
	if stub.requests[0].Auth != want {
		t.Errorf("authorization = %q, want %q", stub.requests[0].Auth, want)
	}
	if got := createdFields(t, stub.requests[0].Body)["issuetype"].(map[string]any)["name"]; got != "Task" {
		t.Errorf("issuetype.name = %v, want Task", got)
	}
}

func TestCreateIssueCustomFieldPrecedence(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue": writeJSON(201, `{"key":"APP-3"}`),
	})
	custom, err := ParseCustomField(`summary=From custom field`)
	if err != nil {
		t.Fatalf("ParseCustomField returned error: %v", err)
	}
	custom["description"] = json.RawMessage(`"overridden"`)

	if _, err := stub.jira(baseJiraConfig()).CreateIssue(context.Background(), CreateIssueParams{
		Summary:     "From flag",
		Description: "From description flag",
		CustomField: custom,
	}); err != nil {
		t.Fatalf("CreateIssue returned error: %v", err)
	}

	fields := createdFields(t, stub.requests[0].Body)
	if got := fields["summary"]; got != "From custom field" {
		t.Errorf("summary = %v, want the custom field to win", got)
	}
	description, ok := fields["description"].(map[string]any)
	if !ok {
		t.Fatalf("description = %v, want --description to win", fields["description"])
	}
	if got := description["type"]; got != "doc" {
		t.Errorf("description.type = %v, want doc", got)
	}
}

func TestCreateIssueFailsOnErrorEnvelope(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue": writeJSON(200, `{"errorMessages":["Field required"],"errors":{"summary":"nope"}}`),
	})

	_, err := stub.jira(baseJiraConfig()).CreateIssue(context.Background(), CreateIssueParams{Summary: "Broken"})

	if err == nil {
		t.Fatal("CreateIssue succeeded, want error from the 200 error envelope")
	}
	if !strings.Contains(err.Error(), "Field required") {
		t.Errorf("error = %q, want it to mention the envelope message", err.Error())
	}
}

func TestChangeIssueStatusMatchesCaseInsensitively(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"GET /rest/api/3/issue/APP-1/transitions":  writeJSON(200, `{"transitions":[{"id":"31","name":"In Review"}]}`),
		"POST /rest/api/3/issue/APP-1/transitions": writeJSON(204, ``),
	})

	if err := stub.jira(baseJiraConfig()).ChangeIssueStatus(context.Background(), "APP-1", "in review"); err != nil {
		t.Fatalf("ChangeIssueStatus returned error: %v", err)
	}

	transition := jsonBody(t, stub.requests[1].Body)["transition"].(map[string]any)
	if got := transition["id"]; got != "31" {
		t.Errorf("transition.id = %v, want 31", got)
	}
}

func TestChangeIssueStatusUnknownStatus(t *testing.T) {
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1/transitions": writeJSON(200, `{"transitions":[{"id":"31","name":"In Review"},{"id":"41","name":"Done"}]}`),
	})

	err := stub.jira(baseJiraConfig()).ChangeIssueStatus(context.Background(), "APP-1", "Shipped")

	if err == nil {
		t.Fatal("ChangeIssueStatus succeeded, want error")
	}
	want := `Status "Shipped" not found. Available statuses: In Review, Done`
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestIssueRawIsNotReencoded(t *testing.T) {
	body := `{"zeta":1,"id":1234567890123456789,"fields":{"summary":"a < b"}}`
	stub := newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1": writeJSON(200, body),
	})

	raw, err := stub.jira(baseJiraConfig()).IssueRaw(context.Background(), "APP-1")
	if err != nil {
		t.Fatalf("IssueRaw returned error: %v", err)
	}

	if string(raw) != body {
		t.Errorf("raw = %q, want %q", raw, body)
	}
}

func TestAttachmentsAndDownload(t *testing.T) {
	var stub *jiraStub
	stub = newJiraStub(t, map[string]func(http.ResponseWriter){
		"/rest/api/3/issue/APP-1": func(w http.ResponseWriter) {
			writeJSON(200, `{"fields":{"attachment":[{"id":"1","filename":"spec.pdf","size":4,"mimeType":"application/pdf","content":"`+stub.server.URL+`/attachment/1","created":"2026-01-01"}]}}`)(w)
		},
		"/attachment/1": func(w http.ResponseWriter) {
			w.Write([]byte("PDF!"))
		},
	})
	client := stub.jira(baseJiraConfig())

	attachments, err := client.Attachments(context.Background(), "APP-1")
	if err != nil {
		t.Fatalf("Attachments returned error: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("got %d attachments, want 1", len(attachments))
	}
	if attachments[0].Filename != "spec.pdf" || attachments[0].Size != 4 {
		t.Errorf("attachment = %+v, want spec.pdf of size 4", attachments[0])
	}

	outputPath := filepath.Join(t.TempDir(), "spec.pdf")
	if err := client.DownloadAttachment(context.Background(), attachments[0].Content, outputPath); err != nil {
		t.Fatalf("DownloadAttachment returned error: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(content) != "PDF!" {
		t.Errorf("downloaded content = %q, want %q", content, "PDF!")
	}
}
