package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestExcludedSearchPhrases(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeExcludedSearchPhrases))
	assert.NoError(t, err)
	err = internal.WriteUint32(buf, 3)
	assert.NoError(t, err)
	err = internal.WriteString(buf, "test 1")
	assert.NoError(t, err)
	err = internal.WriteString(buf, "test 2")
	assert.NoError(t, err)
	err = internal.WriteString(buf, "test 3")
	assert.NoError(t, err)
	b, err := internal.Pack(buf.Bytes())
	assert.NoError(t, err)

	excludedSearchPhrases := new(ExcludedSearchPhrases)
	err = excludedSearchPhrases.Deserialize(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, 3, len(excludedSearchPhrases.Phrases))
	assert.Equal(t, "test 1", excludedSearchPhrases.Phrases[0])
	assert.Equal(t, "test 2", excludedSearchPhrases.Phrases[1])
	assert.Equal(t, "test 3", excludedSearchPhrases.Phrases[2])
	assert.Equal(t, 3, len(excludedSearchPhrases.Phrases))
}
