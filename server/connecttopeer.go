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

const ConnectToPeerCode soul.ServerCode = 18

type ConnectToPeer struct {
	Username       string
	Type           soul.ConnectionType
	IP             net.IP
	Port           int
	Token          uint32
	Privileged     bool
	ObfuscatedPort int
}

func (c ConnectToPeer) Serialize(token uint32, username string, connType soul.ConnectionType) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(ConnectToPeerCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, token)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, string(connType))
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

func (c *ConnectToPeer) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 18
	if err != nil {
		return err
	}

	if code != uint32(ConnectToPeerCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ConnectToPeerCode, code))
	}

	username, err := internal.ReadString(reader)
	if err != nil {
		return err
	}

	c.Username = username

	connType, err := internal.ReadString(reader)
	if err != nil {
		return err
	}

	c.Type = soul.ConnectionType(connType)

	ip, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	c.IP = internal.ReadIP(ip)

	c.Port, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	c.Token, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	c.Privileged, err = internal.ReadBool(reader)
	if err != nil {
		return err
	}

	_, err = internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	c.ObfuscatedPort, err = internal.ReadUint32ToInt(reader)
	if err != nil {
		return err
	}

	return nil
}
