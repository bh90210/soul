package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserInfoRequest(t *testing.T) {
	t.Parallel()

	uir := new(UserInfoRequest)
	message, err := uir.Serialize(uir)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(UserInfoRequest)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, uir, des)
}
