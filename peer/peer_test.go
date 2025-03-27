package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	t.Parallel()

	expected := new(PeerInit)
	expected.Username = "test"
	expected.ConnectionType = "test"
	message, err := expected.Serialize(expected)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	r, size, code, err := Read(CodeInit(0), bytes.NewReader(message), false)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 21, size)
	assert.Equal(t, CodePeerInit, code)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	i, err := Write(buf, &PeerInit{Username: "test", ConnectionType: ConnectionType}, true)
	assert.NoError(t, err)
	assert.Equal(t, 26, i)

	r, size, code, err := Read(CodeInit(0), buf, true)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 18, size)
	assert.Equal(t, CodePeerInit, code)

	login := new(PeerInit)
	login.Deserialize(r)
	assert.Equal(t, "test", login.Username)
	assert.Equal(t, ConnectionType, login.ConnectionType)
}
