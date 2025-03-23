package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadDenied(t *testing.T) {
	t.Parallel()

	ud := new(UploadDenied)
	ud.Filename = "test"
	ud.Reason = ErrBanned
	message, err := ud.Serialize(ud)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(UploadDenied)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, ud, des)
}
