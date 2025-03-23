package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueueUpload(t *testing.T) {
	t.Parallel()

	q := new(QueueUpload)
	q.Filename = "test"
	message, err := q.Serialize(q)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(QueueUpload)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, q, des)
}
