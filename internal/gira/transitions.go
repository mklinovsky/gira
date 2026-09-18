package gira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (j *Jira) Transitions(ctx context.Context, issueKey string) ([]Transition, error) {
	body, err := j.send(ctx, http.MethodGet, j.transitionsURL(issueKey), nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Transitions []Transition `json:"transitions"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("%s: %w", jiraErrorPrefix, err)
	}

	return response.Transitions, nil
}

func (j *Jira) ChangeIssueStatus(ctx context.Context, issueKey, statusName string) error {
	transitions, err := j.Transitions(ctx, issueKey)
	if err != nil {
		return err
	}

	transitionID := ""
	names := make([]string, 0, len(transitions))
	for _, transition := range transitions {
		names = append(names, transition.Name)
		if transitionID == "" && strings.EqualFold(transition.Name, statusName) {
			transitionID = transition.ID
		}
	}

	if transitionID == "" {
		return fmt.Errorf("Status %q not found. Available statuses: %s", statusName, strings.Join(names, ", "))
	}

	payload := map[string]idRef{"transition": {ID: transitionID}}
	_, err = j.send(ctx, http.MethodPost, j.transitionsURL(issueKey), payload)

	return err
}

func (j *Jira) transitionsURL(issueKey string) string {
	return j.config.BaseURL + "/rest/api/3/issue/" + issueKey + "/transitions"
}
