package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadFailed(t *testing.T) {
	t.Parallel()

	uf := new(UploadFailed)
	uf.Filename = "test"
	message, err := uf.Serialize(uf)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(UploadFailed)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, uf, des)
}
