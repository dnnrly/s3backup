package s3

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Config is configuration related to storage in S3
type Config struct {
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`

	ID     string `yaml:"id"`
	Key    string `yaml:"key"`
	Token  string `yaml:"token"`
}

// Store allows you to access your files in an S3 bucket
type Store struct {
	sess   *session.Session
	bucket string
}

// NewStore creates a new Store for you
func NewStore(config Config) (*Store, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.ID,
			config.Key,
			config.Token,
		),
	})

	if err != nil {
		return nil, err
	}

	store := &Store{
		sess:   sess,
		bucket: config.Bucket,
	}

	return store, nil
}

// GetByKey retrieves the data at a certain location in your bucket
func (s *Store) GetByKey(key string) (io.Reader, error) {
	results, err := s3.New(s.sess).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return nil, err
	}
	return buf, nil
}

// Save puts the data at a location in your bucket
func (s *Store) Save(key string, data io.Reader) error {
	return nil
}
