package file

import (
	"bytes"
	"io"

	"github.com/bh90210/soul/internal"
)

// Offset we send this to the uploading peer at the beginning of a ‘F’ connection, to tell them
// how many bytes of the file we’ve previously downloaded.
// If nothing was downloaded, the offset is 0.
type Offset struct {
	Offset uint64
}

// Serialize accepts an offset and returns a message packed as a byte slice.
// The offset is the number of bytes of the file that the peer has previously downloaded.
func (Offset) Serialize(offset uint64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint64(buf, offset)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Deserialize accepts a reader and deserializes the message into the Offset struct.
// The offset is the number of bytes of the file that the peer has previously downloaded.
func (o *Offset) Deserialize(reader io.Reader) (err error) {
	o.Offset, err = internal.ReadUint64(reader)
	return
}
