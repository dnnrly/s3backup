package filemeta

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	// Verbose enables verbose logging in this package
	Verbose = false
)

// Sourcefile represents the metadata for a single backed up file
type Sourcefile struct {
	// Key is the location of this file in a bucket
	Key string `yaml:"key"`
	// Hash is the hashed value of the file contents
	Hash string `yaml:"hash"`
}

// Index holds all of the metadata for files backed up
type Index struct {
	// Files maps the local file location to its metadata
	Files map[string]Sourcefile `yaml:"files"`
}

// NewIndex creates an Index from Yaml
func NewIndex(buf string) (*Index, error) {
	index := &Index{}
	err := yaml.Unmarshal([]byte(buf), index)
	if err != nil {
		return nil, err
	}

	return index, nil
}

// Encode the index data as Yaml
func (i *Index) Encode() (string, error) {
	out, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (local *Index) Diff(remote *Index) *Index {
	diff := &Index{Files: map[string]Sourcefile{}}

	for f, v := range local.Files {
		if _, found := remote.Files[f]; !found {
			diff.Files[f] = v
		}
	}

	return diff
}

type PathHasher func(path string) (string, error)
type PathWalker func(root string, index *Index, hasher PathHasher) filepath.WalkFunc

func FilePathWalker(root string, index *Index, hasher PathHasher) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			doLog("Add in file to index: %s", path)
			key := path
			if root != "" {
				key = fmt.Sprintf("%s/%s", root, path)
			}
			hash, err := hasher(path)
			if err != nil {
				return err
			}
			index.Files[path] = Sourcefile{
				Key:  key,
				Hash: hash,
			}
		}
		return err
	}
}

// FileHasher returns a hash of the contents of a file
func FileHasher(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	h := sha256.Sum256(contents)
	hash := base64.StdEncoding.EncodeToString(h[:])

	return hash, nil
}

// NewIndexFromRoot creates a new Index populated from a filesystem directory
func NewIndexFromRoot(
	bucketRoot,
	path string,
	walker PathWalker,
	hasher PathHasher,
) (*Index, error) {
	i := &Index{
		Files: map[string]Sourcefile{},
	}

	err := filepath.Walk(path, walker(bucketRoot, i, hasher))
	if err != nil {
		return nil, err
	}

	return i, nil
}

func doLog(format string, args ...interface{}) {
	if Verbose {
		log.Printf(format, args...)
	}
}
