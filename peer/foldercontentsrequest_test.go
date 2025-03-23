package peer

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestFolderContentsRequest(t *testing.T) {
	t.Parallel()

	fcr := new(FolderContentsRequest)
	fcr.Token = soul.NewToken()
	fcr.Folder = "test"
	message, err := fcr.Serialize(fcr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(FolderContentsRequest)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, fcr, des)
}
