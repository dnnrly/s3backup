package main

import "gopkg.in/yaml.v2"

// S3Config is configuration related to storage in S3
type S3Config struct {
	Bucket string `yaml:"bucket"`
	ID     string `yaml:"id"`
	Key    string `yaml:"key"`
	Token  string `yaml:"token"`
}

// Config defines the configuration for the whole tool
type Config struct {
	S3 S3Config `yaml:"s3"`
}

// NewConfigFromString generates a config object from the string
func NewConfigFromString(data string) (*Config, error) {
	config := &Config{}
	err := yaml.Unmarshal([]byte(data), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// NewConfigFromFile generates a config object from a file
func NewConfigFromFile(path string) (*S3Config, error) {
	return nil, nil
}
