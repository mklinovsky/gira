package gira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const jiraErrorPrefix = "JIRA API"

type JiraConfig struct {
	BaseURL          string
	APIToken         string
	UserEmail        string
	UserID           string
	ProjectKey       string
	IssueType        string
	SubtaskIssueType string
}

type Jira struct {
	config JiraConfig
	client *http.Client
}

func NewJira(config JiraConfig, client *http.Client) *Jira {
	if client == nil {
		client = http.DefaultClient
	}

	return &Jira{config: config, client: client}
}

func (j *Jira) BrowseURL(issueKey string) string {
	return j.config.BaseURL + "/browse/" + issueKey
}

func (j *Jira) newRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	var reader *strings.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}

	var req *http.Request
	var err error
	if reader == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, reader)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", jiraErrorPrefix, err)
	}

	req.SetBasicAuth(j.config.UserEmail, j.config.APIToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

type jiraErrorEnvelope struct {
	ErrorMessages []string        `json:"errorMessages"`
	Errors        json.RawMessage `json:"errors"`
}

func jiraEnvelopeError(body []byte) error {
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}

	var envelope jiraErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil
	}

	hasErrors := len(envelope.Errors) > 0 &&
		string(envelope.Errors) != "null" &&
		string(envelope.Errors) != "{}"
	if len(envelope.ErrorMessages) == 0 && !hasErrors {
		return nil
	}

	return errors.New(strings.TrimSpace(strings.Join(envelope.ErrorMessages, ",") + " " + string(envelope.Errors)))
}

func (j *Jira) send(ctx context.Context, method, url string, payload any) ([]byte, error) {
	var body []byte
	if payload != nil {
		encoded, err := marshalJSON(payload)
		if err != nil {
			return nil, err
		}
		body = encoded
	}

	req, err := j.newRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	response, err := doRaw(j.client, req, jiraErrorPrefix)
	if err != nil {
		return nil, err
	}

	if err := jiraEnvelopeError(response); err != nil {
		return nil, err
	}

	return response, nil
}
