package main

import (
	"testing"

	"github.com/russross/blackfriday/v2"
)

func TestParseFile(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		markdown    string
		expected    *Issue
		expectError bool
	}{
		{
			name: "ValidMarkdown",
			markdown: `
# Epic 1
Summary of epic 1

## Story 1.1
Summary of story 1.1

### Subtask 1.1.1
Summary of subtask 1.1.1

### Subtask 1.1.2
Summary of subtask 1.1.2

## Story 1.2
Summary of story 1.2
`,
			expected: &Issue{
				Type:    Epic,
				Summary: "Summary of epic 1",
				Children: []*Issue{
					{
						Type:     Story,
						Summary:  "Summary of story 1.1",
						Children: []*Issue{
							{
								Type:    Subtask,
								Summary: "Summary of subtask 1.1.1",
							},
							{
								Type:    Subtask,
								Summary: "Summary of subtask 1.1.2",
							},
						},
					},
					{
						Type:    Story,
						Summary: "Summary of story 1.2",
					},
				},
			},
			expectError: false,
		},
		{
			name: "InvalidMarkdown",
			markdown: `
# Epic 1
Summary of epic 1

## Story 1.1
Summary of story 1.1
`,
			expected:    nil,
			expectError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issue, err := parseFileFromString(tc.markdown)
			if (err != nil) != tc.expectError {
				t.Errorf("Unexpected error: %v", err)
			}

			if err == nil && !issuesEqual(issue, tc.expected) {
				t.Errorf("Parsed issue does not match expected")
			}
		})
	}
}

// parseFileFromString is a helper function to parse a markdown string instead of a file.
func parseFileFromString(content string) (*Issue, error) {
	ast := markdownParser().Parse([]byte(content))

	return parseMarkdownAST(ast)
}

// markdownParser is a helper function to create a blackfriday Markdown parser with common extensions.
func markdownParser() *blackfriday.Parser {
	return blackfriday.New(
		blackfriday.WithExtensions(
			blackfriday.CommonExtensions,
		),
	)
}

// issuesEqual is a helper function to compare two issues for equality.
func issuesEqual(issue1, issue2 *Issue) bool {
	if issue1 == nil && issue2 == nil {
		return true
	}
	if issue1 == nil || issue2 == nil {
		return false
	}

	if issue1.Type != issue2.Type || issue1.Summary != issue2.Summary {
		return false
	}

	if len(issue1.Children) != len(issue2.Children) {
		return false
	}

	for i := range issue1.Children {
		if !issuesEqual(issue1.Children[i], issue2.Children[i]) {
			return false
		}
	}

	return true
}
