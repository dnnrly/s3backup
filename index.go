package s3backup

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// Verbose enables verbose logging in this package
	Verbose = false
)

const (
	indexFile = ".index.yaml"
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

// CopyIndex creates an Index from Yaml
func CopyIndex(from *Index) *Index {
	to := &Index{
		Files: map[string]Sourcefile{},
	}
	for k, v := range from.Files {
		to.Add(k, v)
	}

	return to
}

// Encode the index data as Yaml
func (i *Index) Encode() (string, error) {
	out, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// Add a single source file to the index
func (i *Index) Add(f string, src Sourcefile) {
	i.Files[f] = src
}

// GetNextN gets any N items from the index
func (i *Index) GetNextN(n int) *Index {
	result := &Index{
		Files: map[string]Sourcefile{},
	}

	if len(i.Files) == 0 {
		return result
	}

	for f, src := range i.Files {
		if n == 0 {
			break
		}

		result.Add(f, src)
		n--
	}

	return result
}

// Diff finds all entries in this Index that do not exist or are different from
// the remote entry.
func (local *Index) Diff(remote *Index) *Index {
	diff := &Index{Files: map[string]Sourcefile{}}

	for f, v := range local.Files {
		if _, found := remote.Files[f]; !found {
			log.Printf("Found missing file %s\n", f)
			diff.Files[f] = v
		} else {
			if v.Hash != remote.Files[f].Hash {
				log.Printf("Found updated file %s\n", f)
				diff.Files[f] = v
			}
		}
	}

	return diff
}

// PathHasher is a function that will hash the file at 'path' location
type PathHasher func(path string) (string, error)

// PathWalker is a function that can walk a directory tree and populate the Index
// that is passed in
type PathWalker func(root string, index *Index, hasher PathHasher) filepath.WalkFunc

// FilePathWalker is a PathWalker that accesses files on the disk when walking a
// directory tree
func FilePathWalker(root string, index *Index, hasher PathHasher) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			doLog("Add in file to index: %s", path)
			key := normalisePath(path)
			if root != "" {
				key = fmt.Sprintf("%s/%s", root, key)
			}
			hash, errHash := hasher(path)
			if err != nil {
				return errHash
			}
			index.Files[path] = Sourcefile{
				Key:  key,
				Hash: hash,
			}
		}
		return err
	}
}

func normalisePath(path string) string {
	parts := strings.Split(path, "\\")
	return strings.Join(parts, "/")
}

// FileHasher returns a hash of the contents of a file
func FileHasher(path string) (string, error) {
	contents, err := ioutil.ReadFile(filepath.Clean(path))
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

// FileRepository allows you to access files in your remote location
type FileRepository interface {
	// GetByKey retrieves the data at a certain location in your store
	GetByKey(key string) (io.Reader, error)
	// Save puts the data at a location in your store
	Save(key string, data io.Reader) error
}

// FileGetter allows you to get the contents of a file
type FileGetter func(p string) io.ReadCloser

// IndexStore allows you to persist indexed objects
type IndexStore interface {
	// Save an indexed object to the specified location
	Save(key string, data io.Reader) error
}

// UploadDifferences will upload the files that are missing from the remote index
func UploadDifferences(localIndex, remoteIndex *Index, interval int, store IndexStore, getFile FileGetter) error {
	diff := localIndex.Diff(remoteIndex)
	toUpload := CopyIndex(remoteIndex)
	count := 0

	uploadIndex := func() error {
		r, _ := toUpload.Encode()
		doLog("Uploading index as %s\n", indexFile)
		err := store.Save(indexFile, bytes.NewBufferString(r))
		if err != nil {
			return err
		}
		return nil
	}

	for p, srcFile := range diff.Files {
		r := getFile(p)
		defer func() {
			_ = r.Close()
		}()

		doLog("Uploading %s as %s\n", p, srcFile.Key)
		err := store.Save(srcFile.Key, r)
		if err != nil {
			return err
		}

		count++
		toUpload.Add(p, srcFile)
		if count == interval {
			err := uploadIndex()
			if err != nil {
				return err
			}

			count = 0
		}
	}

	err := uploadIndex()
	if err != nil {
		return err
	}

	return nil
}
