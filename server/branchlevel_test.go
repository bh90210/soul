package server

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
	"github.com/stretchr/testify/assert"
)

func Testbranchlevel(t *testing.T) {
	t.Parallel()

	branchLevel := new(BranchLevel)
	message, err := branchLevel.Serialize(1)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	_, err = buf.Write(message)
	assert.NoError(t, err)

	n, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, uint32(5), n)

	code, err := internal.ReadUint32(buf) // code 66
	assert.NoError(t, err)
	assert.Equal(t, AcceptChildrenCode, soul.CodeServer(code))

	m, err := internal.ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, m)
}
