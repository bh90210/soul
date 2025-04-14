package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestBranchRoot(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	n, err := Write(buf, &BranchRoot{
		Root: "test",
	})

	assert.NoError(t, err)
	assert.Equal(t, 16, n)

	size, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, 12, int(size))

	code, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, uint32(CodeBranchRoot), code)

	root, err := internal.ReadString(buf)
	assert.NoError(t, err)
	assert.Equal(t, "test", root)
}
