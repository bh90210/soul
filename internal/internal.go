// Package internal holds low level functions that are meant to be used internally only.
package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/bh90210/soul"
)

const (
	// VersionMajor is a known accepted client version accepted by the SoulSeek network.
	VersionMajor uint32 = 160
	// VersionMinor is a known accepted client version accepted by the SoulSeek network.
	VersionMinor uint32 = 1
)

// CodeServer messages are used by clients to interface with the server over a connection (TCP).
type CodeServer int

// CodePeerInit This message is sent to initiate a direct connection to another peer.
type CodePeerInit int

// CodePeer messages are sent to peers over a P connection (TCP). Only a single
// active connection to a peer is allowed.
type CodePeer int

// CodeDistributed messages are sent to peers over a D connection (TCP), and are used
// for the distributed search network. Only a single active connection to a peer is allowed.
type CodeDistributed int

// Code is an interface that is used to determine the type of the message.
// It is used in the MessageRead function to determine the type of the message
// that is being read from the connection.
type Code interface {
	CodeServer | CodePeerInit | CodePeer | CodeDistributed
}

// MessageRead reads a message from the connection. It reads the size of the message
// and the code of the message. It then reads the message from the connection and
// returns the message, the size of the message, the code of the message and an error.
func MessageRead[C Code](c C, connection net.Conn) (*bytes.Buffer, int, C, error) {
	message := new(bytes.Buffer)

	// We need to make two reads from the connection to determine the code of the message.
	// Because we need these information down the line we TeeRead them to the message.
	// Note that there is no "message header" in the protocol, we just read the size and code
	// from the "head" of the packet.
	messageHeader := io.TeeReader(connection, message)

	// Read the size of the packet.
	size, err := ReadUint32(messageHeader)
	if err != nil {
		return nil, 0, 0, err
	}

	// Read the code of the message.
	var code C
	var readAlready int
	switch any(c).(type) {
	case CodePeerInit, CodeDistributed:
		c, err := ReadUint8(messageHeader)
		if err != nil {
			return nil, 0, 0, err
		}

		code = C(c)

		readAlready = 1

	case CodeServer, CodePeer:
		c, err := ReadUint32(messageHeader)
		if err != nil {
			return nil, 0, 0, err
		}

		code = C(c)

		readAlready = 4
	}

	// Now we simply copy a packet size read from the connection to the message buffer.
	// This continues writing the message buffer from where the TeeReader left off.
	// The size of the actual message read needs -4 to account for the packet
	// size and code reads that happened above.
	n, err := io.CopyN(message, connection, int64(size)-int64(readAlready))
	if err != nil {
		return nil, 0, 0, err
	}

	// Conversely, we need to add 4 to the size of the total read to account for the
	// size and code reads that are missing from CopyN.
	n += int64(readAlready)

	if int64(size) != n {
		return nil, 0, 0, soul.ErrDifferentPacketSize
	}

	return message, int(size), code, nil
}

// MessageWrite writes a message to the connection. It writes the message to the connection
// and returns the number of bytes written and an error.
func MessageWrite(connection net.Conn, message []byte) (int, error) {
	n, err := connection.Write(message)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// Pack packs the data into a byte slice. It writes the size of the data and the data
// into a buffer and returns the buffer as a byte slice.
func Pack(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ReadUint8 reads a uint8 value from the buffer.
func ReadUint8(buf io.Reader) (uint8, error) {
	var val uint8
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return val, err
}

// WriteUint8 writes a uint8 value to the buffer.
func WriteUint8(buf io.Writer, val uint8) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

// ReadInt32 reads an int32 value from the buffer.
func ReadInt32(buf io.Reader) (int32, error) {
	var val int32
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return val, err
}

// WriteInt32 writes an int32 value to the buffer.
func WriteInt32(buf io.Writer, val int32) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

// ReadInt32ToInt reads an int32 value from the buffer and converts it to an int.
func ReadInt32ToInt(buf io.Reader) (int, error) {
	val, err := ReadInt32(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return int(val), err
}

// ReadUint32 reads a uint32 value from the buffer.
func ReadUint32(buf io.Reader) (uint32, error) {
	var val uint32
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return val, err
}

// WriteUint32 writes a uint32 value to the buffer.
func WriteUint32(buf io.Writer, val uint32) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

// ReadUint32ToInt reads a uint32 value from the buffer and converts it to an int.
func ReadUint32ToInt(buf io.Reader) (int, error) {
	v, err := ReadUint32(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return int(v), err
}

// ReadUint32ToToken reads a uint32 value from the buffer and converts it to a Token.
func ReadUint32ToToken(buf io.Reader) (soul.Token, error) {
	v, err := ReadUint32(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return soul.Token(v), err
}

// ReadUint64 reads a uint64 value from the buffer.
func ReadUint64(reader io.Reader) (uint64, error) {
	var val uint64
	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return val, err
}

// WriteUint64 writes a uint64 value to the buffer.
func WriteUint64(buf io.Writer, val uint64) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

// ReadUint64ToInt reads a uint64 value from the buffer and converts it to an int.
func ReadUint64ToInt(buf io.Reader) (int, error) {
	val, err := ReadUint64(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}

	return int(val), err
}

// NewString creates a byte slice from a string.
func NewString(val string) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.WriteString(val)
	if err != nil {
		return nil, err
	}

	return Pack(buf.Bytes())
}

// ReadString reads a string from the buffer.
func ReadString(reader io.Reader) (string, error) {
	size, err := ReadUint32(reader)
	if err != nil {
		return "", err
	}

	buf := make([]byte, size)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// WriteString writes a string to the buffer.
func WriteString(buf io.Writer, val string) error {
	c, err := NewString(val)
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, c)
	if err != nil {
		return err
	}

	return nil
}

// ReadBool reads a bool value from the buffer.
func ReadBool(reader io.Reader) (bool, error) {
	var val uint8

	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	return val == 1, err
}

// WriteBool writes a bool value to the buffer.
func WriteBool(buf io.Writer, val bool) error {
	var b uint8
	if val {
		b = 1
	}

	return binary.Write(buf, binary.LittleEndian, b)
}

// ReadBytes reads a byte slice from the buffer.
func ReadBytes(reader io.Reader) (buf []byte, err error) {
	size, err := ReadUint32(reader)
	if err != nil {
		return nil, err
	}

	buf = make([]byte, size)
	_, err = io.ReadFull(reader, buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	return buf, err
}

// WriteBytes writes a byte slice to the buffer.
func WriteBytes(buf io.Writer, content []byte) error {
	err := binary.Write(buf, binary.LittleEndian, uint32(len(content)))
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, content)
	if err != nil {
		return err
	}

	return nil
}

// ReadIP reads an IP address in uint32 and returns net.IP..
func ReadIP(val uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, val) // TODO: check why endianess is different than the rest.

	return ip
}
