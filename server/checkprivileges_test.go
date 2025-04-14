package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

// TODO: test via integration.
func TestCheckPrivileges(t *testing.T) {
	t.Parallel()

	cp := new(CheckPrivileges)
	message, err := cp.Serialize(cp)
	assert.NoError(t, err)
	assert.Equal(t, 8, len(message))

	buf := new(bytes.Buffer)
	err = internal.WriteUint32(buf, uint32(CodeCheckPrivileges))
	assert.NoError(t, err)

	err = internal.WriteUint32(buf, uint32(10))
	assert.NoError(t, err)

	message, err = internal.Pack(buf.Bytes())
	assert.NoError(t, err)
	assert.NotNil(t, message)

	cp = new(CheckPrivileges)
	err = cp.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, 10, cp.TimeLeft)
}
