package peer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSharedFileListResponse(t *testing.T) {
	t.Parallel()

	sfr := new(SharedFileListResponse)

	dir := []Directory{
		{
			Name: "test",
			Files: []File{
				File{
					Name:      "test",
					Size:      100,
					Extension: "flac",
					Attributes: []Attribute{
						{
							Code:  FileAttributeType(1),
							Value: 1,
						},
					},
				},
			},
		},
	}

	sfr.Directories = dir
	sfr.PrivateDirectories = dir

	message, err := sfr.Serialize(sfr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(SharedFileListResponse)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, sfr, des)
}
