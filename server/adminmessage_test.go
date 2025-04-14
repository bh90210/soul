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
	err := internal.WriteUint32(buf, uint32(CodeAdminMessage))
	assert.NoError(t, err)
	err = internal.WriteString(buf, "test")
	assert.NoError(t, err)
	b, err := internal.Pack(buf.Bytes())
	assert.NoError(t, err)

	adminMessage := new(AdminMessage)
	err = adminMessage.Deserialize(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, "test", adminMessage.Message)
}
