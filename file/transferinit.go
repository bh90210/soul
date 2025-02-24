package file

import (
	"bytes"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// TransferInit we send this to a peer via a ‘F’ connection to tell them that we want to start uploading a file.
// The token is the same as the one previously included in the TransferRequest peer message.
type TransferInit struct {
	Token soul.Token
}

// Serialize accepts a token and returns a message packed as a byte slice.
func (t TransferInit) Serialize(token soul.Token) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Deserialize accepts a reader and deserializes the message into the TransferInit struct.
func (t *TransferInit) Deserialize(reader io.Reader) (err error) {
	t.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return
	}

	return
}
