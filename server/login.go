package server

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"

	soul "github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// Code Login.
const LoginCode soul.ServerCode = 1

// Response is the message we get from the server when trying to login.
// It can either be a success or a failure.
type Login struct {
	Greet string
	IP    net.IP
	Sum   string
}

// Serialize accepts a username and password. It will create a new byte array (buffer)
// and serialize the username and password into the buffer. It will then calculate
// the sum of the username and password and append it to the buffer. Finally, it will
// append the major and minor version of the protocol to the buffer and return the
// buffer as a byte array.
func (l Login) Serialize(username string, password string) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := internal.WriteUint32(buf, uint32(LoginCode))
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, username)
	if err != nil {
		return nil, err
	}

	err = internal.WriteString(buf, password)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, soul.MajorVersion)
	if err != nil {
		return nil, err
	}

	s, err := sum(username, password)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, s)
	if err != nil {
		return nil, err
	}

	err = internal.WriteUint32(buf, soul.MinorVersion)
	if err != nil {
		return nil, err
	}

	return internal.Pack(buf.Bytes())
}

// ErrInvalidUsername username is longer than 30 characters or contains invalid characters (non-ASCII)
var ErrInvalidUsername = errors.New("INVALIDUSERNAME")

// ErrInvalidPass Password for existing user is incorrect.
var ErrInvalidPass = errors.New("INVALIDPASS")

// ErrInvalidVersion Client version is outdated.
var ErrInvalidVersion = errors.New("INVALIDVERSION")

// Deserialize accepts a reader (from the TCP connection) and reads the response from the server.
// It returns a Response struct containing either a Success or a Failure.
// Consumers of Deserialize must check if the response is OK before proceeding
// as contents f Response are pointers and can be nil.
func (l *Login) Deserialize(reader io.Reader) error {
	_, err := internal.ReadUint32(reader) // size
	if err != nil {
		return err
	}

	code, err := internal.ReadUint32(reader) // code 1
	if err != nil {
		return err
	}

	if code != uint32(LoginCode) {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", LoginCode, code))
	}

	success, err := internal.ReadBool(reader)
	if err != nil {
		return err
	}

	if !success {
		errMessage, err := internal.ReadString(reader)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		switch errMessage {
		case "INVALIDUSERNAME":
			return ErrInvalidUsername

		case "INVALIDPASS":
			return ErrInvalidPass

		case "INVALIDVERSION":
			return ErrInvalidVersion
		}

		// This is not suppose to happen thus we are not
		// dedicating a new var Err for it.
		return fmt.Errorf("unknown login failure: %s", errMessage)
	}

	l.Greet, err = internal.ReadString(reader)
	if err != nil {
		return err
	}

	ip, err := internal.ReadUint32(reader)
	if err != nil {
		return err
	}

	l.IP = internal.ReadIP(ip)

	l.Sum, err = internal.ReadString(reader)
	return err
}

func sum(username string, password string) ([]byte, error) {
	sum := md5.Sum([]byte(username + password))
	return internal.NewString(hex.EncodeToString(sum[:]))
}
