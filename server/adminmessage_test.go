package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestAdminMessage(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	internal.WriteUint32(buf, uint32(CodeAdminMessage))
	internal.WriteString(buf, "test")
	b, _ := internal.Pack(buf.Bytes())

	buf = new(bytes.Buffer)
	buf.Write(b)

	adminMessage := new(AdminMessage)
	err := adminMessage.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, "test", adminMessage.Message)
}
