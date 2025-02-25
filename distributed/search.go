package distributed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// Search code 3 request that arrives through the distributed network.
// We transmit the search request to our child peers.
type Search struct {
	Token    soul.Token
	Username string
	Query    string
}

// Serialize accepts a token, username, and query and returns a message packed as a byte slice.
func (d Search) Serialize(token soul.Token, username, query string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(CodeSearch))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(0))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, query)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize accepts a reader and deserializes the message into the Search struct.
func (d *Search) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 3
	if err != nil {
		return err
	}

	if code != uint8(CodeSearch) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeSearch, code))
	}

	_, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	d.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	d.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	d.Query, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
