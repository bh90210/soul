package server

import (
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/distributed"
	"github.com/bh90210/soul/internal"
)

const CodeEmbeddedMessage Code = 93

// EmbeddedMessage code 93, the server sends us an embedded distributed message. The only
// type of distributed message sent at present is DistribSearch (distributed code 3).
// If we receive such a message, we are a branch root in the distributed network,
// and we distribute the embedded message (not the unpacked distributed message) to our child peers.
type EmbeddedMessage struct {
	Code    distributed.Code
	Message []byte
}

func (e *EmbeddedMessage) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 93
	if err != nil {
		return err
	}

	if code != uint32(CodeEmbeddedMessage) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodeEmbeddedMessage, code))
	}

	embeddedCode, err := internal.ReadUint8(reader)
	if err != nil {
		return err
	}

	e.Code = distributed.Code(embeddedCode)

	e.Message, err = internal.ReadBytes(reader)
	if err != nil {
		return err
	}

	return nil

}
