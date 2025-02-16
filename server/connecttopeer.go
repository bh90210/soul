package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/bh90210/soul"
)

const ConnectToPeerCode Code = 18

type ConnectToPeer struct {
	Username       string
	Type           soul.ConnectionType
	IP             net.IP
	Port           int
	Token          int
	Privileged     bool
	ObfuscatedPort int
}

func (c ConnectToPeer) Serialize(token int, username string, connType soul.ConnectionType) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := soul.WriteUint32(buf, uint32(ConnectToPeerCode))
	if err != nil {
		return nil, err
	}

	err = soul.WriteUint32(buf, uint32(token))
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = soul.WriteString(buf, string(connType))
	if err != nil {
		return nil, err
	}

	return soul.Pack(buf.Bytes())
}

func (c *ConnectToPeer) Deserialize(reader io.Reader) error {
	_, err := soul.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := soul.ReadUint32(reader) // code 18
	if err != nil {
		return err
	}

	if code != uint32(ConnectToPeerCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", ConnectToPeerCode, code))
	}

	username, err := soul.ReadString(reader)
	if err != nil {
		return err
	}

	c.Username = username

	connType, err := soul.ReadString(reader)
	if err != nil {
		return err
	}

	c.Type = soul.ConnectionType(connType)

	ip, err := soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	c.IP = soul.ReadIP(ip)

	c.Port, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	c.Token, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	c.Privileged, err = soul.ReadBool(reader)
	if err != nil {
		return err
	}

	_, err = soul.ReadUint32(reader)
	if err != nil {
		return err
	}

	c.ObfuscatedPort, err = soul.ReadInt(reader)
	if err != nil {
		return err
	}

	return nil
}
