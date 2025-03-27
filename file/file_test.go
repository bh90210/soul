package file

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	expected := new(TransferInit)
	expected.Token = soul.NewToken()

	buf := new(bytes.Buffer)
	i, err := Write(buf, expected)
	assert.NoError(t, err)
	assert.Equal(t, 4, i)

	actual := new(TransferInit)
	err = actual.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected.Token, actual.Token)

	off := new(Offset)
	off.Offset = 123
	i, err = Write(buf, off)
	assert.NoError(t, err)
	assert.Equal(t, 8, i)

	actualOff := new(Offset)
	err = actualOff.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, off.Offset, actualOff.Offset)
}
