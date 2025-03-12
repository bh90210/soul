package distributed

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	t.Parallel()

	token := soul.NewToken()

	search := new(Search)
	message, err := search.Serialize(token, "test", "query")
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	_, err = buf.Write(message)
	assert.NoError(t, err)

	err = search.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, token, search.Token)
	assert.Equal(t, "test", search.Username)
	assert.Equal(t, "query", search.Query)
}
