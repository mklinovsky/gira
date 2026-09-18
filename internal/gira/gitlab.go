package gira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const gitlabErrorPrefix = "GitLab API"

type GitlabConfig struct {
	BaseURL   string
	APIToken  string
	ProjectID string
	UserID    string
}

type Gitlab struct {
	config GitlabConfig
	client *http.Client
}

func NewGitlab(config GitlabConfig, client *http.Client) *Gitlab {
	if client == nil {
		client = http.DefaultClient
	}

	return &Gitlab{config: config, client: client}
}

type CreateMergeRequestParams struct {
	SourceBranch string
	TargetBranch string
	Title        string
	Labels       string
}

func (g *Gitlab) CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (string, error) {
	payload := map[string]any{
		"source_branch":        params.SourceBranch,
		"target_branch":        params.TargetBranch,
		"title":                params.Title,
		"assignee_id":          g.config.UserID,
		"remove_source_branch": true,
	}
	if params.Labels != "" {
		payload["labels"] = params.Labels
	}

	body, err := g.send(ctx, http.MethodPost, g.mergeRequestsURL(""), payload)
	if err != nil {
		return "", err
	}

	var created struct {
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		return "", fmt.Errorf("%s: %w", gitlabErrorPrefix, err)
	}

	return created.WebURL, nil
}

func (g *Gitlab) MergeRequestRaw(ctx context.Context, mergeRequestID string) ([]byte, error) {
	return g.send(ctx, http.MethodGet, g.mergeRequestsURL(mergeRequestID), nil)
}

func (g *Gitlab) MergeRequestTitle(ctx context.Context, mergeRequestID string) (string, error) {
	body, err := g.MergeRequestRaw(ctx, mergeRequestID)
	if err != nil {
		return "", err
	}

	var mergeRequest struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(body, &mergeRequest); err != nil {
		return "", fmt.Errorf("%s: %w", gitlabErrorPrefix, err)
	}

	return mergeRequest.Title, nil
}

func (g *Gitlab) Merge(ctx context.Context, mergeRequestID string, deleteSourceBranch bool) error {
	payload := map[string]bool{"should_remove_source_branch": deleteSourceBranch}
	_, err := g.send(ctx, http.MethodPut, g.mergeRequestsURL(mergeRequestID)+"/merge", payload)

	return err
}

func (g *Gitlab) mergeRequestsURL(mergeRequestID string) string {
	url := g.config.BaseURL + "/api/v4/projects/" + g.config.ProjectID + "/merge_requests"
	if mergeRequestID != "" {
		url += "/" + mergeRequestID
	}

	return url
}

func (g *Gitlab) send(ctx context.Context, method, url string, payload any) ([]byte, error) {
	var reader *strings.Reader
	if payload != nil {
		encoded, err := marshalJSON(payload)
		if err != nil {
			return nil, err
		}
		reader = strings.NewReader(string(encoded))
	}

	var req *http.Request
	var err error
	if reader == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, reader)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", gitlabErrorPrefix, err)
	}

	req.Header.Set("PRIVATE-TOKEN", g.config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	return doRaw(g.client, req, gitlabErrorPrefix)
}
