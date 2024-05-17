package main

import (
    "bufio"
    "bytes"
    "errors"
    "io/ioutil"
    "regexp"
    "strings"

    "github.com/russross/blackfriday/v2"
)

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
