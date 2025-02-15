package getpeeraddress

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	"github.com/bh90210/soul"
)

// Code GetPeerAddress.
const Code soul.UInt = 3

// Serialize accepts a username and returns a serialized byte array.
func Serialize(username string) []byte {
	buf := new(bytes.Buffer)
	soul.WriteUInt(buf, Code)

	binary.Write(buf, binary.LittleEndian, soul.NewString(username))

	return soul.Pack(buf.Bytes())
}

// Response is the message we get from the server when trying to get a peer's address.
type Response struct {
	Username       string
	IP             net.IP
	Port           soul.UInt
	ObfuscatedPort soul.UInt
}

func Deserialize(reader io.Reader) *Response {
	soul.ReadUInt(reader) // size
	soul.ReadUInt(reader) // code 3

	username := soul.ReadString(reader)
	ip := soul.ReadIP(soul.ReadUInt(reader))
	port := soul.ReadUInt(reader)
	soul.ReadUInt(reader)
	obfuscatedPort := soul.ReadUInt(reader)

	return &Response{username, ip, port, obfuscatedPort}
}
