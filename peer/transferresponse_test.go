package peer

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestTransferResponse(t *testing.T) {
	t.Parallel()

	tr := new(TransferResponse)
	tr.Allowed = false
	tr.Token = soul.NewToken()
	tr.Reason = ErrBanned
	message, err := tr.Serialize(tr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(TransferResponse)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, tr, des)
}
