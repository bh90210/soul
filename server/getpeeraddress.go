package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
)

// Code GetPeerAddress.
const GetPeerAddressCode soul.UInt = 3

// Response is the message we get from the server when trying to get a peer's address.
type GetPeerAddress struct {
	Username       string
	IP             net.IP
	Port           soul.UInt
	ObfuscatedPort soul.UInt
}

// Serialize accepts a username and returns a serialized byte array.
func (g GetPeerAddress) Serialize(username string) ([]byte, error) {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, GetPeerAddressCode)

	err := binary.Write(buf, binary.LittleEndian, soul.NewString(username))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes()), nil
}

func (g *GetPeerAddress) Deserialize(reader io.Reader) error {
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 3
	if code != GetPeerAddressCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", GetPeerAddressCode, code))

	}

	g.Username = soul.ReadString(reader)
	g.IP = soul.ReadIP(soul.ReadUInt(reader))
	g.Port = soul.ReadUInt(reader)
	soul.ReadUInt(reader)
	g.ObfuscatedPort = soul.ReadUInt(reader)

	return nil
}
