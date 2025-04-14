package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestCantConnectToPeer(t *testing.T) {
	t.Parallel()

	token := soul.NewToken()
	ccp := new(CantConnectToPeer)
	ccp.Token = token
	ccp.Username = "test"

	message, err := ccp.Serialize(ccp)
	assert.NoError(t, err)
	assert.Equal(t, 20, len(message))

	err = ccp.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, token, ccp.Token)
	assert.Equal(t, "test", ccp.Username)
}
