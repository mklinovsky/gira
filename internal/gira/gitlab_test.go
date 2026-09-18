package gira

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newGitlabStub(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) (*Gitlab, *[]recordedRequest) {
	t.Helper()

	requests := &[]recordedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*requests = append(*requests, recordedRequest{
			Method: r.Method,
			Path:   r.URL.Path,
			Auth:   r.Header.Get("PRIVATE-TOKEN"),
			Body:   string(body),
		})
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client := NewGitlab(GitlabConfig{
		BaseURL:   server.URL,
		APIToken:  "glpat-token",
		ProjectID: "42",
		UserID:    "7",
	}, server.Client())

	return client, requests
}

func TestCreateMergeRequest(t *testing.T) {
	client, requests := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(201, `{"web_url":"https://gitlab.example.com/mr/1"}`)(w)
	})

	url, err := client.CreateMergeRequest(context.Background(), CreateMergeRequestParams{
		SourceBranch: "APP-1-feature",
		TargetBranch: "main",
		Title:        "APP-1 Feature",
		Labels:       "backend,urgent",
	})
	if err != nil {
		t.Fatalf("CreateMergeRequest returned error: %v", err)
	}

	if want := "https://gitlab.example.com/mr/1"; url != want {
		t.Errorf("url = %q, want %q", url, want)
	}

	recorded := (*requests)[0]
	if want := "/api/v4/projects/42/merge_requests"; recorded.Path != want {
		t.Errorf("path = %q, want %q", recorded.Path, want)
	}
	if recorded.Auth != "glpat-token" {
		t.Errorf("PRIVATE-TOKEN = %q, want %q", recorded.Auth, "glpat-token")
	}

	payload := jsonBody(t, recorded.Body)
	for key, want := range map[string]any{
		"source_branch":        "APP-1-feature",
		"target_branch":        "main",
		"title":                "APP-1 Feature",
		"assignee_id":          "7",
		"remove_source_branch": true,
		"labels":               "backend,urgent",
	} {
		if payload[key] != want {
			t.Errorf("payload[%q] = %v, want %v", key, payload[key], want)
		}
	}
}

func TestCreateMergeRequestOmitsEmptyLabels(t *testing.T) {
	client, requests := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(201, `{"web_url":"https://gitlab.example.com/mr/2"}`)(w)
	})

	if _, err := client.CreateMergeRequest(context.Background(), CreateMergeRequestParams{
		SourceBranch: "branch",
		TargetBranch: "main",
		Title:        "Title",
	}); err != nil {
		t.Fatalf("CreateMergeRequest returned error: %v", err)
	}

	if _, present := jsonBody(t, (*requests)[0].Body)["labels"]; present {
		t.Error("payload contains labels, want it omitted")
	}
}

func TestMergeRequestRawIsNotReencoded(t *testing.T) {
	body := `{"zeta":1,"iid":9007199254740993,"title":"APP-1 Feature"}`
	client, _ := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(200, body)(w)
	})

	raw, err := client.MergeRequestRaw(context.Background(), "1")
	if err != nil {
		t.Fatalf("MergeRequestRaw returned error: %v", err)
	}
	if string(raw) != body {
		t.Errorf("raw = %q, want %q", raw, body)
	}

	title, err := client.MergeRequestTitle(context.Background(), "1")
	if err != nil {
		t.Fatalf("MergeRequestTitle returned error: %v", err)
	}
	if want := "APP-1 Feature"; title != want {
		t.Errorf("title = %q, want %q", title, want)
	}
}

func TestMergeAcceptsNoContent(t *testing.T) {
	client, requests := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	if err := client.Merge(context.Background(), "1", true); err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}

	recorded := (*requests)[0]
	if want := "/api/v4/projects/42/merge_requests/1/merge"; recorded.Path != want {
		t.Errorf("path = %q, want %q", recorded.Path, want)
	}
	if recorded.Method != http.MethodPut {
		t.Errorf("method = %q, want PUT", recorded.Method)
	}
	if got := jsonBody(t, recorded.Body)["should_remove_source_branch"]; got != true {
		t.Errorf("should_remove_source_branch = %v, want true", got)
	}
}

func TestMergeRequestWrapsHTTPError(t *testing.T) {
	client, _ := newGitlabStub(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := client.MergeRequestRaw(context.Background(), "404")

	if err == nil {
		t.Fatal("MergeRequestRaw succeeded, want error")
	}
	if got := err.Error(); got[:len("GitLab API: 404 Not Found ")] != "GitLab API: 404 Not Found " {
		t.Errorf("error = %q, want the GitLab API 404 prefix", got)
	}
}
