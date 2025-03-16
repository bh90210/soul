package server

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

const CodePossibleParents Code = 102

// PossibleParents code 102, the server send us a list of max 10 possible distributed
// parents to connect to. Messages of this type are sent to us at regular intervals,
// until we tell the server we don’t need more possible parents with a HaveNoParent message.
// The received list always contains users whose upload speed is higher than our own.
// If we have the highest upload speed on the server, we become a branch root, and start
// receiving SearchRequest messages directly from the server.
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

	if code != uint32(CodePossibleParents) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", CodePossibleParents, code))
	}

	parents, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	for range int(parents) {
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
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		p.Parents = append(p.Parents, parent)
	}

	return err
}
