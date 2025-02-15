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
)

// Code Login.
const LoginCode soul.UInt = 1

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
func (l Login) Serialize(username string, password string) []byte {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, LoginCode)

	binary.Write(buf, binary.LittleEndian, soul.NewString(username))
	binary.Write(buf, binary.LittleEndian, soul.NewString(password))
	binary.Write(buf, binary.LittleEndian, soul.MajorVersion)
	binary.Write(buf, binary.LittleEndian, sum(username, password))
	binary.Write(buf, binary.LittleEndian, soul.MinorVersion)

	return soul.Pack(buf.Bytes())
}

// Deserialize accepts a reader (from the TCP connection) and reads the response from the server.
// It returns a Response struct containing either a Success or a Failure.
// Consumers of Deserialize must check if the response is OK before proceeding
// as contents f Response are pointers and can be nil.
func (l *Login) Deserialize(reader io.Reader) error {
	soul.ReadUInt(reader)         // size
	code := soul.ReadUInt(reader) // code 1
	if code != LoginCode {
		return errors.Join(soul.ErrMismatchingCodes,
			fmt.Errorf("expected code %d, got %d", LoginCode, code))
	}

	success := soul.ReadBool(reader)
	if !success {
		return readFailure(reader)
	}

	l.readSuccess(reader)

	return nil
}

func (l *Login) readSuccess(reader io.Reader) {
	l.Greet = soul.ReadString(reader)
	l.IP = soul.ReadIP(soul.ReadUInt(reader))
	l.Sum = soul.ReadString(reader)
}

// ErrLoginFailureInvalidUsername username is longer than 30 characters or contains invalid characters (non-ASCII)
var ErrLoginFailureInvalidUsername = errors.New("INVALIDUSERNAME")

// ErrLoginFailureInvalidPass Password for existing user is incorrect.
var ErrLoginFailureInvalidPass = errors.New("INVALIDPASS")

// ErrLoginFailureInvalidVersion Client version is outdated.
var ErrLoginFailureInvalidVersion = errors.New("INVALIDVERSION")

func readFailure(reader io.Reader) error {
	switch soul.ReadString(reader) {
	case "INVALIDUSERNAME":
		return ErrLoginFailureInvalidUsername

	case "INVALIDPASS":
		return ErrLoginFailureInvalidPass

	case "INVALIDVERSION":
		return ErrLoginFailureInvalidVersion
	}

	// This is not suppose to happen thus we are not
	// dedicating a new var Err for it.
	return errors.New("unknown login failure: " + soul.ReadString(reader))
}

func sum(username string, password string) soul.String {
	sum := md5.Sum([]byte(username + password))
	return soul.NewString(hex.EncodeToString(sum[:]))
}
