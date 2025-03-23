package peer

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestFileSearchResponse(t *testing.T) {
	t.Parallel()

	fsr := new(FileSearchResponse)
	fsr.Username = "test"
	fsr.Token = soul.NewToken()
	file := File{
		Name:      "test",
		Size:      100,
		Extension: "flac",
		Attributes: []Attribute{
			{
				Code:  FileAttributeType(1),
				Value: 1,
			},
		},
	}

	fsr.Results = []File{file}
	fsr.FreeSlot = true
	fsr.AverageSpeed = 1
	fsr.Queue = 1
	fsr.PrivateResults = []File{file}
	message, err := fsr.Serialize(fsr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(FileSearchResponse)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, fsr, des)
}
