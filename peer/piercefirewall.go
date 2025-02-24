package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const PierceFirewallCode soul.PeerInitCode = 0

type PierceFirewall struct {
	Token soul.Token
}

func (p PierceFirewall) Serialize(token soul.Token) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(PierceFirewallCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (p *PierceFirewall) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 0
	if err != nil {
		return err
	}

	if code != uint8(PierceFirewallCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %v, got %v", PierceFirewallCode, code))
	}

	p.Token, err = internal.ReadUint32ToToken(reader)
	if err != nil {
		return err
	}

	return nil
}
