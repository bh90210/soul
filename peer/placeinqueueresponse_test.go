package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceInQueueResponse(t *testing.T) {
	t.Parallel()

	pqr := new(PlaceInQueueResponse)
	pqr.Filename = "test"
	pqr.Place = 1
	message, err := pqr.Serialize(pqr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(PlaceInQueueResponse)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, pqr, des)
}
