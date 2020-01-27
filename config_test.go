package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigFromString_DiscoversParseError(t *testing.T) {
	data := `
not vali
`
	config, err := NewConfigFromString(data)

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestNewConfigFromString(t *testing.T) {
	data := `
extra:
  - A
  - B
s3:
  bucket: "bucket-01"
  id: "ID-1"
  key: "Key-1"
  token: "Token-1"
`
	expected := &Config{
		S3: S3Config{
			Bucket: "bucket-01",
			ID:     "ID-1",
			Key:    "Key-1",
			Token:  "Token-1",
		},
	}

	config, err := NewConfigFromString(data)

	assert.NoError(t, err)
	assert.Equal(t, expected, config)
}
