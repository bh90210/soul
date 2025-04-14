package server

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangePassword(t *testing.T) {
	t.Parallel()

	cp := new(ChangePassword)
	cp.Pass = "test"

	message, err := cp.Serialize(cp)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	cp = new(ChangePassword)
	err = cp.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, "test", cp.Pass)
}
