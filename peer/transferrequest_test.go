package peer

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestTransferRequest(t *testing.T) {
	t.Parallel()

	tr := new(TransferRequest)
	tr.Filename = "test"
	tr.FileSize = 100
	tr.Token = soul.NewToken()
	tr.Direction = UploadToPeer
	message, err := tr.Serialize(tr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(TransferRequest)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, tr, des)
}
