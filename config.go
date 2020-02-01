package s3backup

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/dnnrly/s3backup/s3"
	"gopkg.in/yaml.v2"
)

// Config defines the configuration for the whole tool
type Config struct {
	S3 s3.Config `yaml:"s3"`
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
func NewConfigFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read contents of config file: %w", err)
	}

	return NewConfigFromString(string(b))
}
