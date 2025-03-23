package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePierceFirewall CodeInit = 0

// PierceFirewall code 0 message is sent in response to an indirect connection
// request from another user. If the message goes through to the user, the connection
// is ready. The token is taken from the ConnectToPeer server message.
type PierceFirewall struct {
	Token soul.Token
}

// Serialize accepts a PierceFirewall and returns a message packed as a byte slice.
func (p *PierceFirewall) Serialize(message *PierceFirewall) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(CodePierceFirewall))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(message.Token))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// Deserialize populates a PierceFirewall with the data in the provided reader.
func (p *PierceFirewall) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 0
	if err != nil {
		return err
	}

	if code != uint8(CodePierceFirewall) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %v, got %v", CodePierceFirewall, code))
	}

	p.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	return nil
}
