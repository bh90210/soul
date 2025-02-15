package login

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"

	soul "github.com/bh90210/soul"
)

// Code Login.
const Code soul.UInt = 1

// Serialize accepts a username and password. It will create a new byte array (buffer)
// and serialize the username and password into the buffer. It will then calculate
// the sum of the username and password and append it to the buffer. Finally, it will
// append the major and minor version of the protocol to the buffer and return the
// buffer as a byte array.
func Serialize(username string, password string) []byte {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, Code)

	binary.Write(buf, binary.LittleEndian, soul.NewString(username))
	binary.Write(buf, binary.LittleEndian, soul.NewString(password))
	binary.Write(buf, binary.LittleEndian, soul.MajorVersion)
	binary.Write(buf, binary.LittleEndian, sum(username, password))
	binary.Write(buf, binary.LittleEndian, soul.MinorVersion)

	return soul.Pack(buf.Bytes())
}

// Response is the message we get from the server when trying to login.
// It can either be a success or a failure.
type Response struct {
	Greet string
	IP    net.IP
	Sum   string
}

// Deserialize accepts a reader (from the TCP connection) and reads the response from the server.
// It returns a Response struct containing either a Success or a Failure.
// Consumers of Deserialize must check if the response is OK before proceeding
// as contents f Response are pointers and can be nil.
func Deserialize(reader io.Reader) (*Response, error) {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 1

	success := soul.ReadBool(reader)
	if !success {
		return nil, readFailure(reader)
	}

	return readSuccess(reader), nil
}

func readSuccess(reader io.Reader) *Response {
	greet := soul.ReadString(reader)
	ip := soul.ReadIP(soul.ReadUInt(reader))
	sum := soul.ReadString(reader)

	return &Response{greet, ip, sum}
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
