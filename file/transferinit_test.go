package file

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestTransferInit(t *testing.T) {
	t.Parallel()

	token := soul.NewToken()

	transferInit := new(TransferInit)
	transferInit.Token = token
	r, err := transferInit.Serialize(transferInit)
	assert.NoError(t, err)
	assert.NotNil(t, r)

	buf := new(bytes.Buffer)
	_, err = buf.Write(r)
	assert.NoError(t, err)

	err = transferInit.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, token, transferInit.Token)
}
