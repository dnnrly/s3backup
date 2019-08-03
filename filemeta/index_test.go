package filemeta

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIndex(t *testing.T) {
	buf := `files:
  a/b/c:
    key: root/a/b/c
  d/e:
    key: root/d/e
  f:
    key: root/f
`

	index, err := NewIndex(buf)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(index.Files))
	assert.Equal(t, "root/a/b/c", index.Files["a/b/c"].Key)
	assert.Equal(t, "root/d/e", index.Files["d/e"].Key)
	assert.Equal(t, "root/f", index.Files["f"].Key)
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
				Key: "a/b/c",
			},
		},
	}

	out, _ := i.Encode()
	assert.Equal(t, `files:
    1/2/3:
        key: a/b/c
`, out)
}

func TestIndexDifference(t *testing.T) {
	local := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a"},
			"2": Sourcefile{Key: "b"},
			"3": Sourcefile{Key: "c"},
			"4": Sourcefile{Key: "d"},
		},
	}
	remote := &Index{
		Files: map[string]Sourcefile{
			"1": Sourcefile{Key: "a"},
			"2": Sourcefile{Key: "b"},
			"4": Sourcefile{Key: "d"},
			"5": Sourcefile{Key: "e"},
		},
	}

	diff := local.Diff(remote)

	assert.Equal(t, 1, len(diff.Files))
	assert.Equal(t, "c", diff.Files["3"].Key)
}
