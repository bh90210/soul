package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func TestBranchLevel(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	n, err := Write(buf, &BranchLevel{
		Level: 2,
	})

	assert.NoError(t, err)
	assert.Equal(t, 12, n)

	size, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, 8, int(size))

	code, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, uint32(CodeBranchLevel), code)

	level, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(level))
}
