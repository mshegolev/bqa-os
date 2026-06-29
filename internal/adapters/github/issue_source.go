package github

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type IssueJSONSource struct {
	Path string
}

func (s IssueJSONSource) LoadTeamIssue(ctx context.Context, ref ports.TeamIssueRef) (ports.TeamIssue, error) {
	select {
	case <-ctx.Done():
		return ports.TeamIssue{}, ctx.Err()
	default:
	}

	path := s.Path
	if path == "" {
		path = ref.JSONPath
	}
	if path == "" {
		return ports.TeamIssue{}, errors.New("GitHub issue JSON path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ports.TeamIssue{}, err
	}

	var raw struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return ports.TeamIssue{}, err
	}

	number := raw.Number
	if number == 0 {
		number = ref.Number
	}

	labels := make([]ports.TeamIssueLabel, 0, len(raw.Labels))
	for _, label := range raw.Labels {
		if label.Name != "" {
			labels = append(labels, ports.TeamIssueLabel{Name: label.Name})
		}
	}

	return ports.TeamIssue{
		Number: number,
		Title:  raw.Title,
		Body:   raw.Body,
		Labels: labels,
	}, nil
}
