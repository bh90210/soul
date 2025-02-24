package file

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffset(t *testing.T) {
	t.Parallel()

	offset := new(Offset)
	r, err := offset.Serialize(1)
	assert.NoError(t, err)
	assert.NotNil(t, r)

	buf := new(bytes.Buffer)
	_, err = buf.Write(r)
	assert.NoError(t, err)

	err = offset.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), offset.Offset)
}
