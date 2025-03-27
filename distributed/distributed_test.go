package distributed

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	t.Parallel()

	expected := new(BranchLevel)
	expected.Level = 123
	message, err := expected.Serialize(expected)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	r, size, code, err := Read(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 5, size)
	assert.Equal(t, CodeBranchLevel, code)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	expected := new(BranchLevel)
	expected.Level = 123
	message, err := expected.Serialize(expected)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	i, err := Write(buf, &BranchLevel{Level: 123})
	assert.NoError(t, err)
	assert.Equal(t, 9, i)
	assert.Equal(t, message, buf.Bytes())
}
