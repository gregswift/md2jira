package main

import (
    "io/ioutil"
    "path/filepath"

    "gopkg.in/yaml.v2"
)

// Config represents the configuration structure.
type Config struct {
    Endpoint string `yaml:"endpoint"`
    User     string `yaml:"user"`
    Token    string `yaml:"token"`
}

// loadConfig loads the configuration from the YAML config file.
func loadConfig(path string) (*Config, error) {
    content, err := ioutil.ReadFile(filepath.Clean(path))
    if err != nil {
        return nil, err
    }

    var config Config
    err = yaml.Unmarshal(content, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}
