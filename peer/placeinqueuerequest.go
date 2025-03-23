package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePlaceInQueueRequest Code = 51

// PlaceInQueueRequest code 51 message is sent when asking for
// the upload queue placement of a file.
type PlaceInQueueRequest struct {
	Filename string
}

// Serialize accepts a PlaceInQueueRequest and returns a message packed as a byte slice.
func (PlaceInQueueRequest) Serialize(message *PlaceInQueueRequest) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodePlaceInQueueRequest))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Filename)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize populates a PlaceInQueueRequest with the data in the provided reader.
func (p *PlaceInQueueRequest) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 51
	if err != nil {
		return err
	}

	if code != uint32(CodePlaceInQueueRequest) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePlaceInQueueRequest, code))
	}

	p.Filename, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	return nil
}
