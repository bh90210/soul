package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfoResponse(t *testing.T) {
	t.Parallel()

	uir := new(UserInfoResponse)
	uir.Description = "test"
	uir.Picture = []byte("test")
	uir.TotalUpload = 1
	uir.QueueSize = 1
	uir.FreeSlots = true
	message, err := uir.Serialize(uir)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(UserInfoResponse)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, uir, des)
}
