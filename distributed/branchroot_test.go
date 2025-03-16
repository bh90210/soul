package distributed

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBranchRoot(t *testing.T) {
	t.Parallel()

	branchRoot := new(BranchRoot)
	branchRoot.Root = "test"
	message, err := branchRoot.Serialize(branchRoot)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	buf := new(bytes.Buffer)
	_, err = buf.Write(message)
	assert.NoError(t, err)

	err = branchRoot.Deserialize(buf)
	assert.NoError(t, err)
	assert.Equal(t, "test", branchRoot.Root)
}
