package main

// IssueType represents the type of a Jira issue.
type IssueType int

const (
    Epic IssueType = iota
    Story
    Subtask
)

// Issue represents a Jira issue, including its type, summary, children, and labels.
type Issue struct {
    Type     IssueType
    Summary  string
    Children []*Issue
    Labels   []string
    Fields   map[string]string
}
