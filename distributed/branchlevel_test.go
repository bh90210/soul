package distributed

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBranchLevel(t *testing.T) {
	t.Parallel()

	branchLevel := new(BranchLevel)
	message, err := branchLevel.Serialize(1)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	_, err = buf.Write(message)
	assert.NoError(t, err)

	err = branchLevel.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), branchLevel.Level)
}
