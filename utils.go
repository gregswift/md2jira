package main

import (
    "errors"
    "strings"
)

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
