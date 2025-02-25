package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// PeerInitCode code 1.
type PeerInit struct {
	RemoteUsername string
	ConnectionType soul.ConnectionType
}

func (PeerInit) Serialize(ownUsername string, connType soul.ConnectionType) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint8(buf, uint8(CodePeerInit))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, ownUsername)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, string(connType))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, uint32(0)) // unknown
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (p *PeerInit) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint8(reader) // code 1
	if err != nil {
		return err
	}

	if code != uint8(CodePeerInit) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePeerInit, code))
	}

	p.RemoteUsername, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	connType, err := internal.ReadString(reader)
	if err != nil {
		return err
	}

	p.ConnectionType = soul.ConnectionType(connType)

	_, err = internal.ReadUint32(reader) // unknown
	if err != nil {
		return err
	}

	return nil
}
