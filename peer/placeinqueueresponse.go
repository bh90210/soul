package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePlaceInQueueResponse Code = 44

// PlaceInQueueResponse code 44 peer replies with the upload queue placement
// of the requested file.
type PlaceInQueueResponse struct {
	Filename string
	Place    uint32
}

func (p *PlaceInQueueResponse) Serialize(message *PlaceInQueueResponse) ([]byte, error) {
	buf := new(bytes.Buffer)

	err := internal.WriteUint32(buf, uint32(CodePlaceInQueueResponse))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, message.Filename)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, message.Place)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (p *PlaceInQueueResponse) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 44
	if err != nil {
		return err
	}

	if code != uint32(CodePlaceInQueueResponse) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePlaceInQueueResponse, code))
	}

	p.Filename, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	p.Place, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	return nil
}
