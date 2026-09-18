package gira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateIssueParams struct {
	Summary     string
	IssueType   string
	ParentKey   string
	Description string
	AssignToMe  bool
	CustomField map[string]json.RawMessage
}

type CreatedIssue struct {
	Key string
	URL string
}

type keyRef struct {
	Key string `json:"key"`
}

type nameRef struct {
	Name string `json:"name"`
}

type idRef struct {
	ID string `json:"id"`
}

type adfText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type adfParagraph struct {
	Type    string    `json:"type"`
	Content []adfText `json:"content"`
}

type adfDocument struct {
	Type    string         `json:"type"`
	Version int            `json:"version"`
	Content []adfParagraph `json:"content"`
}

type issueFields struct {
	Fields map[string]json.RawMessage `json:"fields"`
}

type jiraIssueType struct {
	HierarchyLevel *int `json:"hierarchyLevel"`
	Subtask        bool `json:"subtask"`
}

type jiraIssue struct {
	Fields struct {
		IssueType  *jiraIssueType `json:"issuetype"`
		Attachment []Attachment   `json:"attachment"`
	} `json:"fields"`
}

func (j *Jira) CreateIssue(ctx context.Context, params CreateIssueParams) (CreatedIssue, error) {
	issueType, err := j.resolveIssueType(ctx, params.IssueType, params.ParentKey)
	if err != nil {
		return CreatedIssue{}, err
	}

	fields := map[string]json.RawMessage{}
	if err := putField(fields, "project", keyRef{Key: j.config.ProjectKey}); err != nil {
		return CreatedIssue{}, err
	}
	if err := putField(fields, "summary", params.Summary); err != nil {
		return CreatedIssue{}, err
	}
	if err := putField(fields, "issuetype", nameRef{Name: issueType}); err != nil {
		return CreatedIssue{}, err
	}
	if params.AssignToMe {
		if err := putField(fields, "assignee", idRef{ID: j.config.UserID}); err != nil {
			return CreatedIssue{}, err
		}
	}
	if params.ParentKey != "" {
		if err := putField(fields, "parent", keyRef{Key: params.ParentKey}); err != nil {
			return CreatedIssue{}, err
		}
	}
	for key, value := range params.CustomField {
		fields[key] = value
	}
	if params.Description != "" {
		if err := putField(fields, "description", descriptionDocument(params.Description)); err != nil {
			return CreatedIssue{}, err
		}
	}

	body, err := j.send(ctx, http.MethodPost, j.config.BaseURL+"/rest/api/3/issue", issueFields{Fields: fields})
	if err != nil {
		return CreatedIssue{}, err
	}

	var created struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		return CreatedIssue{}, fmt.Errorf("%s: %w", jiraErrorPrefix, err)
	}

	return CreatedIssue{Key: created.Key, URL: j.BrowseURL(created.Key)}, nil
}

func (j *Jira) UpdateIssue(ctx context.Context, issueKey string, customField map[string]json.RawMessage) error {
	if customField == nil {
		customField = map[string]json.RawMessage{}
	}

	_, err := j.send(
		ctx,
		http.MethodPut,
		j.config.BaseURL+"/rest/api/3/issue/"+issueKey,
		issueFields{Fields: customField},
	)

	return err
}

func (j *Jira) IssueRaw(ctx context.Context, issueKey string) ([]byte, error) {
	return j.send(ctx, http.MethodGet, j.config.BaseURL+"/rest/api/3/issue/"+issueKey, nil)
}

func (j *Jira) resolveIssueType(ctx context.Context, issueType, parentKey string) (string, error) {
	if issueType != "" {
		return issueType, nil
	}

	if parentKey == "" {
		return j.standardIssueType(), nil
	}

	body, err := j.IssueRaw(ctx, parentKey)
	if err != nil {
		return "", err
	}

	var parent jiraIssue
	if err := json.Unmarshal(body, &parent); err != nil {
		return "", fmt.Errorf("%s: %w", jiraErrorPrefix, err)
	}

	parentType := parent.Fields.IssueType
	if parentType == nil {
		return "", fmt.Errorf("Could not determine issue type for parent %s.", parentKey)
	}
	if parentType.Subtask {
		return "", fmt.Errorf("Cannot create child issue under subtask %s.", parentKey)
	}
	if parentType.HierarchyLevel != nil && *parentType.HierarchyLevel > 0 {
		return j.standardIssueType(), nil
	}
	if j.config.SubtaskIssueType == "" {
		return "", fmt.Errorf("Jira setting subtaskIssueType is required when parent issue requires subtask creation.")
	}

	return j.config.SubtaskIssueType, nil
}

func (j *Jira) standardIssueType() string {
	if j.config.IssueType == "" {
		return "Task"
	}

	return j.config.IssueType
}

func descriptionDocument(description string) adfDocument {
	return adfDocument{
		Type:    "doc",
		Version: 1,
		Content: []adfParagraph{{
			Type:    "paragraph",
			Content: []adfText{{Type: "text", Text: description}},
		}},
	}
}

func putField(fields map[string]json.RawMessage, key string, value any) error {
	encoded, err := marshalJSON(value)
	if err != nil {
		return err
	}
	fields[key] = encoded

	return nil
}
