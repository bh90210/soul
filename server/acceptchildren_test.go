package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestAcceptChildren(t *testing.T) {
	t.Parallel()

	acceptChildren := new(AcceptChildren)
	acceptChildren.Accept = true
	message, err := acceptChildren.Serialize(acceptChildren)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	_, err = buf.Write(message)
	assert.NoError(t, err)

	n, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, uint32(5), n)

	code, err := internal.ReadUint32(buf) // code 66
	assert.NoError(t, err)
	assert.Equal(t, CodeAcceptChildren, Code(code))

	m, err := internal.ReadBool(buf)
	assert.NoError(t, err)
	assert.Equal(t, true, m)
}
