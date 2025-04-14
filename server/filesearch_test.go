package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestFileSearch(t *testing.T) {
	t.Parallel()

	token := soul.NewToken()

	fs := new(FileSearch)
	fs.Token = token
	fs.SearchQuery = "test"
	message, err := fs.Serialize(fs)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(message))

	buf := new(bytes.Buffer)
	err = internal.WriteUint32(buf, uint32(CodeFileSearch))
	assert.NoError(t, err)
	err = internal.WriteString(buf, "username")
	assert.NoError(t, err)
	err = internal.WriteUint32(buf, uint32(token))
	assert.NoError(t, err)
	err = internal.WriteString(buf, "query")
	assert.NoError(t, err)
	b, err := internal.Pack(buf.Bytes())
	assert.NoError(t, err)

	fs = new(FileSearch)
	err = fs.Deserialize(bytes.NewBuffer(b))
	assert.NoError(t, err)
	assert.Equal(t, token, fs.Token)
	assert.Equal(t, "query", fs.SearchQuery)
	assert.Equal(t, "username", fs.Username)
}
