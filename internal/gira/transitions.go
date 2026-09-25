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
	To   struct {
		Name string `json:"name"`
	} `json:"to"`
}

// Workflows often name a transition differently from the status it leads to
// ("Start progress" -> "In Progress"), so both are accepted.
func (t Transition) matches(statusName string) bool {
	return strings.EqualFold(t.Name, statusName) || strings.EqualFold(t.To.Name, statusName)
}

func (t Transition) label() string {
	if t.To.Name == "" || strings.EqualFold(t.To.Name, t.Name) {
		return t.Name
	}

	return t.Name + " (" + t.To.Name + ")"
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
	labels := make([]string, 0, len(transitions))
	for _, transition := range transitions {
		labels = append(labels, transition.label())
		if transitionID == "" && transition.matches(statusName) {
			transitionID = transition.ID
		}
	}

	if transitionID == "" {
		return fmt.Errorf("Status %q not found. Available statuses: %s", statusName, strings.Join(labels, ", "))
	}

	payload := map[string]idRef{"transition": {ID: transitionID}}
	_, err = j.send(ctx, http.MethodPost, j.transitionsURL(issueKey), payload)

	return err
}

func (j *Jira) transitionsURL(issueKey string) string {
	return j.config.BaseURL + "/rest/api/3/issue/" + issueKey + "/transitions"
}
