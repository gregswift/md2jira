// main.go
package main

import (
    "bufio"
    "bytes"
    "errors"
    "fmt"
    "io/ioutil"
    "os"
    "strings"

    "github.com/andygrunwald/go-jira"
    "github.com/russross/blackfriday/v2"
)

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

// isJiraTicket checks if a given string is a valid Jira ticket ID.
func isJiraTicket(title string) bool {
    pattern := `^[A-Z]+-\d+$`
    re := regexp.MustCompile(pattern)
    return re.MatchString(title)
}

// parseFile parses a Markdown file and returns an Issue representing the hierarchy of issues in the file.
func parseFile(path string) (*Issue, error) {
    content, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }

    contentStr := string(content)

    markdown := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions))
    ast := markdown.Parse([]byte(contentStr))

    root := &Issue{
        Type:    Epic,
        Summary: "",
    }
    currentIssue := root

    for _, node := range ast.Children {
        if node.Type == blackfriday.Heading {
            level := bytes.Count(node.Literal, []byte{'#'})
            if level > 3 {
                continue
            }

            var summary string
            var labels []string
            scanner := bufio.NewScanner(bytes.NewReader(node.Literal))
            for scanner.Scan() {
                line := scanner.Text()
                if strings.HasPrefix(line, "Labels: ") {
                    labels = strings.Split(strings.TrimPrefix(line, "Labels: "), ",")
                } else {
                    summary += line
                }
            }

            if isJiraTicket(summary) {
                if err := verifyIssueExists(summary); err != nil {
                    return nil, err
                }
                continue
            }

            var issueType IssueType
            switch level {
            case 1:
                issueType = Epic
            case 2:
                issueType = Story
            case 3:
                issueType = Subtask
            }
            newIssue := &Issue{
                Type:     issueType,
                Summary:  strings.TrimSpace(summary),
                Children: nil,
                Labels:   labels,
                Fields:   nil,
            }

            if level == 1 {
                root = newIssue
            } else {
                for len(currentIssue.Children) > 0 && currentIssue.Type >= newIssue.Type {
                    currentIssue = currentIssue.Children[len(currentIssue.Children)-1]
                }
                currentIssue.Children = append(currentIssue.Children, newIssue)
            }

            currentIssue = newIssue
        } else if node.Type == blackfriday.CodeBlock && bytes.HasPrefix(node.Info, []byte("field-settings")) {
            settings := bytes.TrimPrefix(node.Info, []byte("field-settings"))
            fields, err := parseFields(string(settings))
            if err != nil {
                return nil, err
            }
            currentIssue.Fields = fields
        }
    }

    return root, nil
}

// parseFields parses a settings string into a map of field names and values.
func parseFields(settings string) (map[string]string, error) {
    fields := make(map[string]string)
    scanner := bufio.NewScanner(strings.NewReader(settings))
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            return nil, errors.New("invalid field setting: " + line)
        }
        fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
    }
    return fields, nil
}

// verifyIssueExists verifies that a Jira issue with the given key exists.
func verifyIssueExists(key string) error {
    // This function should call the Jira API to verify the issue exists.
    // Here we return nil for simplicity.
    return nil
}

// getFields parses command-line arguments into a map of custom field names and values.
func getFields(args []string) (map[string]string, error) {
    fields := make(map[string]string)
    for _, arg := range args {
        if !strings.HasPrefix(arg, "--field=") {
            continue
        }
        setting := strings.TrimPrefix(arg, "--field=")
        parts := strings.SplitN(setting, "=", 2)
        if len(parts) != 2 {
            return nil, errors.New("invalid field setting: " + setting)
        }
        fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
    }
    return fields, nil
}

// createIssues creates issues in Jira based on an Issue struct, with any additional custom fields specified on the command line.
func createIssues(client *jira.Client, issue *Issue, fields map[string]string) error {
    // This function should create the issues in Jira using the Jira API.
    // Here we print the issue for simplicity.
    fmt.Printf("Creating issue: %s\n", issue.Summary)
    for key, value := range fields {
        fmt.Printf("  %s: %s\n", key, value)
    }
    for _, child := range issue.Children {
        if err := createIssues(client, child, fields); err != nil {
            return err
        }
    }
    return nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: jira-cli <markdown-file> [--field=field_name=value ...]")
        os.Exit(1)
    }

    filepath := os.Args[1]
    fields, err := getFields(os.Args[2:])
    if err != nil {
        fmt.Printf("Error parsing fields: %v\n", err)
        os.Exit(1)
    }

    issue, err := parseFile(filepath)
    if err != nil {
        fmt.Printf("Error parsing file: %v\n", err)
        os.Exit(1)
    }

    jiraClient, err := jira.NewClient(nil, "https://your-jira-instance.com")
    if err != nil {
        fmt.Printf("Error creating Jira client: %v\n", err)
        os.Exit(1)
    }

    err = createIssues(jiraClient, issue, fields)
    if err != nil {
        fmt.Printf("Error creating issues: %v\n", err)
        os.Exit(1)
    }
}
