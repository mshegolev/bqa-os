package ports

import "context"

type TeamIssueRef struct {
	Repo     string
	Number   int
	JSONPath string
}

type TeamIssueLabel struct {
	Name string
}

type TeamIssue struct {
	Number int
	Title  string
	Body   string
	Labels []TeamIssueLabel
}

type TeamIssueSource interface {
	LoadTeamIssue(ctx context.Context, ref TeamIssueRef) (TeamIssue, error)
}
