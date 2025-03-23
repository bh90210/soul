package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerInit(t *testing.T) {
	t.Parallel()

	pi := new(PeerInit)
	pi.Username = "test"
	pi.ConnectionType = ConnectionType
	message, err := pi.Serialize(pi)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(PeerInit)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, pi, des)
}
