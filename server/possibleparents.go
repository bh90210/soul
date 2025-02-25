package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const PossibleParentsCode soul.CodeServer = 102

type PossibleParents struct {
	Parents []Parent
}

type Parent struct {
	Username string
	IP       net.IP
	Port     int
}

func (p *PossibleParents) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 102
	if err != nil {
		return err
	}

	if code != uint32(PossibleParentsCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PossibleParentsCode, code))
	}

	parents, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(parents); i++ {
		var parent Parent

		parent.Username, err = internal.ReadString(reader)
		if err != nil {
			return err
		}

		ip, err := internal.ReadUint32(reader)
		if err != nil {
			return err
		}

		parent.IP = internal.ReadIP(ip)

		parent.Port, err = internal.ReadUint32ToInt(reader)
		if err != nil {
			return err
		}

		p.Parents = append(p.Parents, parent)
	}

	return nil
}
