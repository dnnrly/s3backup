package filemeta

import (
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
