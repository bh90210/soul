package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// Code GetPeerAddress.
const GetPeerAddressCode soul.CodeServer = 3

// Response is the message we get from the server when trying to get a peer's address.
type GetPeerAddress struct {
	Username       string
	IP             net.IP
	Port           int
	ObfuscatedPort int
}

// Serialize accepts a username and returns a serialized byte array.
func (g GetPeerAddress) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(GetPeerAddressCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (g *GetPeerAddress) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 3
	if err != nil {
		return err
	}

	if code != uint32(GetPeerAddressCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetPeerAddressCode, code))
	}

	g.Username, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	ip, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	g.IP = internal.ReadIP(ip)

	g.Port, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	_, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	g.ObfuscatedPort, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
