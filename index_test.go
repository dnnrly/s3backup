package s3backup

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndex(t *testing.T) {
	buf := `files:
  a/b/c:
    key: root/a/b/c
    hash: 123
  d/e:
    key: root/d/e
    hash: 456
  f:
    key: root/f
    hash: 789
`

	index, err := NewIndex(buf)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(index.Files))
	assert.Equal(t, "root/a/b/c", index.Files["a/b/c"].Key)
	assert.Equal(t, "123", index.Files["a/b/c"].Hash)
	assert.Equal(t, "root/d/e", index.Files["d/e"].Key)
	assert.Equal(t, "456", index.Files["d/e"].Hash)
	assert.Equal(t, "root/f", index.Files["f"].Key)
	assert.Equal(t, "789", index.Files["f"].Hash)
}

func TestNewIndex_ParseError(t *testing.T) {
	buf := `files:
  a/b/c: [}
`

	index, err := NewIndex(buf)
	assert.Error(t, err)
	assert.Nil(t, index)
}

func TestEncode(t *testing.T) {
	i := Index{
		Files: map[string]Sourcefile{
			"1/2/3": Sourcefile{
				Key:  "a/b/c",
				Hash: "123",
			},
		},
	}

	out, _ := i.Encode()
	assert.Equal(t, `files:
    1/2/3:
        key: a/b/c
        hash: "123"
`, out)
}

func TestIndexDifference(t *testing.T) {
	local := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "321"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"3": Sourcefile{Key: "c", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
		},
	}
	remote := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "123"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
			"5": Sourcefile{Key: "e", Hash: "123"},
		},
	}

	diff := local.Diff(remote)

	assert.Equal(t, 2, len(diff.Files))
	assert.Equal(t, "321", diff.Files["1"].Hash)
	assert.Equal(t, "c", diff.Files["3"].Key)
}

func TestIndexAdd(t *testing.T) {
	index := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "321"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"3": Sourcefile{Key: "c", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
		},
	}

	index.Add("5", Sourcefile{Key: "e", Hash: "999"})

	assert.Equal(t, 5, len(index.Files))
}

func TestIndexGetNextN_Many(t *testing.T) {
	index := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "321"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"3": Sourcefile{Key: "c", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
		},
	}

	got := index.GetNextN(2)

	assert.Equal(t, 2, len(got.Files))
	for f := range got.Files {
		assert.Equal(t, index.Files[f], got.Files[f])
	}
}

func TestIndexGetNextN_None(t *testing.T) {
	got := &Index{
		Files: map[string]Sourcefile{},
	}
	assert.Equal(t, 0, len(got.GetNextN(1).Files))
}

type mockStore struct {
	Keys      []string
	Values    []string
	FailAfter int
}

func (m *mockStore) Save(key string, data io.Reader) error {
	if len(m.Keys) >= m.FailAfter {
		return errors.New("oops")
	}

	m.Keys = append(m.Keys, key)

	buf := new(bytes.Buffer)
	buf.ReadFrom(data)
	m.Values = append(m.Values, buf.String())

	return nil
}

func TestUploadDifferences(t *testing.T) {
	index := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "321"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"3": Sourcefile{Key: "c", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
			// index
			"5": Sourcefile{Key: "e", Hash: "123"},
			"6": Sourcefile{Key: "f", Hash: "123"},
			"7": Sourcefile{Key: "g", Hash: "123"},
			"8": Sourcefile{Key: "h", Hash: "123"},
			// index
			"9": Sourcefile{Key: "i", Hash: "123"},
			// index
		},
	}

	getter := func(p string) io.ReadCloser {
		s := strings.NewReader("")
		c := ioutil.NopCloser(s)
		return c
	}

	mock := &mockStore{
		Keys:      []string{},
		FailAfter: 99,
	}
	err := UploadDifferences(index, &Index{}, 4, mock, getter)

	assert.Equal(t, 12, len(mock.Keys))
	assert.Equal(t, ".index.yaml", mock.Keys[4])
	assert.Equal(t, ".index.yaml", mock.Keys[9])
	assert.True(t, len(mock.Values[4]) < len(mock.Values[9]))
	assert.Equal(t, ".index.yaml", mock.Keys[11])
	assert.True(t, len(mock.Values[9]) < len(mock.Values[11]))
	assert.NoError(t, err)
}

func TestUploadDifferences_ObjectSaveFails(t *testing.T) {
	index := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "321"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"3": Sourcefile{Key: "c", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
			"5": Sourcefile{Key: "e", Hash: "123"},
			"6": Sourcefile{Key: "f", Hash: "123"},
			"7": Sourcefile{Key: "g", Hash: "123"},
			"8": Sourcefile{Key: "h", Hash: "123"},
			"9": Sourcefile{Key: "i", Hash: "123"},
		},
	}

	getter := func(p string) io.ReadCloser {
		s := strings.NewReader("")
		c := ioutil.NopCloser(s)
		return c
	}

	mock := &mockStore{
		Keys:      []string{},
		FailAfter: 5,
	}
	err := UploadDifferences(index, &Index{}, 4, mock, getter)

	assert.Equal(t, 5, len(mock.Keys))
	assert.Error(t, err)
}

func TestUploadDifferences_IndexSaveFails(t *testing.T) {
	index := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a", Hash: "321"},
			"2": Sourcefile{Key: "b", Hash: "123"},
			"3": Sourcefile{Key: "c", Hash: "123"},
			"4": Sourcefile{Key: "d", Hash: "123"},
			"5": Sourcefile{Key: "d", Hash: "123"},
			"6": Sourcefile{Key: "d", Hash: "123"},
			"7": Sourcefile{Key: "d", Hash: "123"},
			"8": Sourcefile{Key: "d", Hash: "123"},
			"9": Sourcefile{Key: "d", Hash: "123"},
		},
	}

	getter := func(p string) io.ReadCloser {
		s := strings.NewReader("")
		c := ioutil.NopCloser(s)
		return c
	}

	mock := &mockStore{
		Keys:      []string{},
		FailAfter: 9,
	}
	err := UploadDifferences(index, &Index{}, 12, mock, getter)

	assert.Equal(t, 9, len(mock.Keys))
	assert.NotContains(t, ".index.yaml", mock.Keys)
	assert.Error(t, err)
}

func TestNormalisePath(t *testing.T) {
	assert.Equal(t, "a/b/c", normalisePath("a/b/c"))
	assert.Equal(t, "a/b/c", normalisePath("a\\b\\c"))
}
