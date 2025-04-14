package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestEmbeddedMessage(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(CodeEmbeddedMessage))
	assert.NoError(t, err)
	err = internal.WriteUint8(buf, 3)
	assert.NoError(t, err)
	err = internal.WriteBytes(buf, []byte("test"))
	assert.NoError(t, err)
	b, err := internal.Pack(buf.Bytes())
	assert.NoError(t, err)

	embeddedMessage := new(EmbeddedMessage)
	err = embeddedMessage.Deserialize(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, distributed.Code(3), embeddedMessage.Code)
	assert.Equal(t, []byte("test"), embeddedMessage.Message)
}
