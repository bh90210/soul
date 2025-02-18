package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
)

// Code GetPeerAddress.
const GetPeerAddressCode Code = 3

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
	err := soul.WriteUint32(buf, uint32(GetPeerAddressCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (g *GetPeerAddress) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 3
	if err != nil {
		return err
	}

	if code != uint32(GetPeerAddressCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetPeerAddressCode, code))
	}

	g.Username, err = soul.ReadString(reader)
	if err != nil {
		return err
	}

	ip, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	g.IP = soul.ReadIP(ip)

	g.Port, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	_, err = soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	g.ObfuscatedPort, err = soul.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
