package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"

    "github.com/andygrunwald/go-jira"
)

func main() {
    var configPath string
    var jiraURL string

    flag.StringVar(&configPath, "config", filepath.Join(os.Getenv("HOME"), ".jira.d", "config.yml"), "Path to the configuration file")
    flag.StringVar(&jiraURL, "url", "", "Jira URL (overrides config file)")
    flag.Parse()

    if len(flag.Args()) < 1 {
        fmt.Println("Usage: jira-cli [--config=config.yml] [--url=jira-url] <markdown-file> [--field=field_name=value ...]")
        os.Exit(1)
    }

    markdownFile := flag.Args()[0]
    fields, err := getFields(flag.Args()[1:])
    if err != nil {
        fmt.Printf("Error parsing fields: %v\n", err)
        os.Exit(1)
    }

    config, err := loadConfig(configPath)
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        os.Exit(1)
    }

    if jiraURL == "" {
        jiraURL = config.Endpoint
    }

    if jiraURL == "" {
        fmt.Println("Jira URL must be provided either through the config file or the --url flag")
        os.Exit(1)
    }

    jiraEmail := config.User
    jiraAPIToken := config.Token
    if jiraEmail == "" || jiraAPIToken == "" {
        fmt.Println("Jira email and API token must be provided in the config file")
        os.Exit(1)
    }

    tp := jira.BasicAuthTransport{
        Username: jiraEmail,
        APIToken: jiraAPIToken,
    }

    jiraClient, err := jira.NewClient(tp.Client(), jiraURL)
    if err != nil {
        fmt.Printf("Error creating Jira client: %v\n", err)
        os.Exit(1)
    }

    issue, err := parseFile(markdownFile)
    if err != nil {
        fmt.Printf("Error parsing file: %v\n", err)
        os.Exit(1)
    }

    err = createIssues(jiraClient, issue, fields)
    if err != nil {
        fmt.Printf("Error creating issues: %v\n", err)
        os.Exit(1)
    }
}
