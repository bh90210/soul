package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSharedFileListRequest(t *testing.T) {
	t.Parallel()

	sfr := new(SharedFileListRequest)
	message, err := sfr.Serialize(sfr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(SharedFileListRequest)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, sfr, des)
}
