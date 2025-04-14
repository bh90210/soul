package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestCantCreateRoom(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	internal.WriteUint32(buf, uint32(CodeCantCreateRoom))
	internal.WriteString(buf, "test")
	b, err := internal.Pack(buf.Bytes())
	assert.NoError(t, err)

	buf = new(bytes.Buffer)
	buf.Write(b)

	adminMessage := new(CantCreateRoom)
	err = adminMessage.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, "test", adminMessage.Room)
}
