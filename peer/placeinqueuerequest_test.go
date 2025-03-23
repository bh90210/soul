package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceInQueueRequest(t *testing.T) {
	t.Parallel()

	pqr := new(PlaceInQueueRequest)
	pqr.Filename = "test"
	message, err := pqr.Serialize(pqr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(PlaceInQueueRequest)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, pqr, des)
}
