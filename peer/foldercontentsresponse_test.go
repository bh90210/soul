package peer

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestFolderContentsResponse(t *testing.T) {
	t.Parallel()

	fcr := new(FolderContentsResponse)
	fcr.Token = soul.NewToken()
	fcr.Folder = "test"
	folder := Folder{
		Directory: "test",
		Files: []File{
			{
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
	}
	fcr.Folders = []Folder{folder}
	message, err := fcr.Serialize(fcr)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(FolderContentsResponse)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, fcr, des)
}
