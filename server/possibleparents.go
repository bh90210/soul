package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
)

const PossibleParentsCode soul.ServerCode = 102

type PossibleParents struct {
	Parents []Parent
}

type Parent struct {
	Username string
	IP       net.IP
	Port     int
}

func (p *PossibleParents) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 102
	if err != nil {
		return err
	}

	if code != uint32(PossibleParentsCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", PossibleParentsCode, code))
	}

	parents, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	for i := 0; i < int(parents); i++ {
		var parent Parent

		parent.Username, err = soul.ReadString(reader)
		if err != nil {
			return err
		}

		ip, err := soul.ReadUint32(reader)
		if err != nil {
			return err
		}

		parent.IP = soul.ReadIP(ip)

		parent.Port, err = soul.ReadUint32ToInt(reader)
		if err != nil {
			return err
		}

		p.Parents = append(p.Parents, parent)
	}

	return nil
}
