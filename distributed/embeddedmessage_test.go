package distributed

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddedMessage(t *testing.T) {
	t.Parallel()

	embeddedMessage := new(EmbeddedMessage)
	message, err := embeddedMessage.Serialize(0, []byte("test"))
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	_, err = buf.Write(message)
	assert.NoError(t, err)

	err = embeddedMessage.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("test"), embeddedMessage.Message)
}
