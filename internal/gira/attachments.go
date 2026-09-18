package gira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
	Created  string `json:"created"`
}

func (j *Jira) Attachments(ctx context.Context, issueKey string) ([]Attachment, error) {
	body, err := j.IssueRaw(ctx, issueKey)
	if err != nil {
		return nil, err
	}

	var issue jiraIssue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, fmt.Errorf("%s: %w", jiraErrorPrefix, err)
	}

	return issue.Fields.Attachment, nil
}

func (j *Jira) DownloadAttachment(ctx context.Context, contentURL, outputPath string) error {
	req, err := j.newRequest(ctx, http.MethodGet, contentURL, nil)
	if err != nil {
		return err
	}

	body, err := doRaw(j.client, req, jiraErrorPrefix)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, body, 0o644); err != nil {
		return fmt.Errorf("%s: %w", jiraErrorPrefix, err)
	}

	return nil
}
