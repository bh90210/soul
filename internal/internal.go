// Package internal holds low level functions that are meant to be used internally only.
package internal

import (
	"bytes"
	"crypto/rand"
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
func MessageRead[C Code](c C, connection net.Conn, obfuscated bool) (*bytes.Buffer, int, C, error) {
	message := new(bytes.Buffer)

	// We need to make two reads from the connection to determine the code of the message.
	// Because we need these information down the line we TeeRead them to the message.
	// Note that there is no "message header" in the protocol, we just read the size and code
	// from the "head" of the packet.
	messageHeader := io.TeeReader(connection, message)

	// All documentation about obfuscation is coming from the good people of https://aioslsk.readthedocs.io/en/latest/SOULSEEK.html#obfuscation.
	// Read the size of the packet.
	var size uint32
	var err error
	if obfuscated {
		// We need to deobfuscate the size, which is the first packet of the message with a size of 4 bytes.
		deobfuscated, err := teeDeobfuscateN(connection, message, 4)
		if err != nil {
			return nil, 0, 0, err
		}

		size, err = ReadUint32(deobfuscated)
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		size, err = ReadUint32(messageHeader)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Read the code of the message.
	var code C
	var readAlready int64
	switch any(c).(type) {
	case CodePeerInit:
		var c uint8
		if obfuscated {
			deobfuscated, err := teeDeobfuscateN(connection, message, 1)
			if err != nil {
				return nil, 0, 0, err
			}

			c, err = ReadUint8(deobfuscated)
			if err != nil {
				return nil, 0, 0, err
			}

		} else {
			c, err = ReadUint8(messageHeader)
			if err != nil {
				return nil, 0, 0, err
			}
		}

		code = C(c)

		readAlready = 1

	case CodePeer:
		var c uint32
		if obfuscated {
			deobfuscated, err := teeDeobfuscateN(message, connection, 4)
			if err != nil {
				return nil, 0, 0, err
			}

			c, err = ReadUint32(deobfuscated)
			if err != nil {
				return nil, 0, 0, err
			}

		} else {
			c, err = ReadUint32(messageHeader)
			if err != nil {
				return nil, 0, 0, err
			}
		}

		code = C(c)

		readAlready = 4

	case CodeServer:
		c, err := ReadUint32(messageHeader)
		if err != nil {
			return nil, 0, 0, err
		}

		code = C(c)

		readAlready = 4

	case CodeDistributed:
		c, err := ReadUint8(messageHeader)
		if err != nil {
			return nil, 0, 0, err
		}

		code = C(c)

		readAlready = 1
	}

	// Now we simply copy a packet size read from the connection to the message buffer.
	// This continues writing the message buffer from where the TeeReader left off.
	// The size of the actual message read needs -4 to account for the packet
	// size and code reads that happened above.
	var n int64
	if obfuscated {
		n, err = copyDeobfuscateN(message, connection, int64(size)-int64(readAlready))
		if err != nil {
			return nil, 0, 0, err
		}

	} else {
		n, err = io.CopyN(message, connection, int64(size)-int64(readAlready))
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Conversely, we need to add 4 to the size of the total read to account for the
	// size and code reads that are missing from CopyN.
	n += readAlready

	if int64(size) != n {
		return nil, 0, 0, soul.ErrDifferentPacketSize
	}

	return message, int(size), code, nil
}

func copyDeobfuscateN(message io.Writer, connection io.Reader, n int64) (int64, error) {
	deobfuscated, err := deobfuscateN(connection, n)
	if err != nil {
		return 0, err
	}

	return io.CopyN(message, deobfuscated, n)
}

func teeDeobfuscateN(message io.Writer, connection io.Reader, n int64) (io.Reader, error) {
	deobfuscated, err := deobfuscateN(connection, n)
	if err != nil {
		return nil, err
	}

	return io.TeeReader(deobfuscated, message), nil
}

func deobfuscateN(connection io.Reader, n int64) (*bytes.Buffer, error) {
	deobfuscated := new(bytes.Buffer)

	// Key is the first 4 bytes of the message in little-endian.
	key := new(bytes.Buffer)

	// Directly read from the connection to the key buffer.
	i, err := io.CopyN(key, connection, 4)
	if err != nil {
		return nil, err
	}

	if i != 4 {
		return nil, soul.ErrDifferentPacketSize
	}

	var readSoFar int64
	for {
		// Convert it to big-endian integer.
		var bigKey uint32
		err = binary.Read(key, binary.LittleEndian, &bigKey)
		if err != nil {
			return nil, err
		}

		// Shift 31 bits to the right.
		bigKey = (bigKey >> 31) | (bigKey << (32 - 31))

		// Convert back to little-endian byte array.
		rotatedKey := new(bytes.Buffer)
		err = binary.Write(rotatedKey, binary.LittleEndian, bigKey)
		if err != nil {
			return nil, err
		}

		// Read next 4 bytes of the actual message from the connection.
		next4bytes := new(bytes.Buffer)
		i, err := io.CopyN(next4bytes, connection, 4)
		if err != nil {
			return nil, err
		}

		// Track how many bytes we have read so far.
		readSoFar += i

		key.Reset()
		deobfuscated4bytes := new(bytes.Buffer)
		// XOR 4 bytes of the message with the 4 bytes of the rotated key.
		for o := range i {
			b, err := rotatedKey.ReadByte()
			if err != nil {
				return nil, err
			}

			// We construct the new key while reading it for XOR.
			err = key.WriteByte(b)
			if err != nil {
				return nil, err
			}

			b ^= next4bytes.Bytes()[o]
			deobfuscated4bytes.WriteByte(b)
		}

		// Write the deobfuscated 4 bytes to the deobfuscated message buffer.
		deobfuscated.Write(deobfuscated4bytes.Bytes())

		if readSoFar == n {
			break
		}
	}

	return deobfuscated, nil
}

// MessageWrite writes a message to the connection. It writes the message to the connection
// and returns the number of bytes written and an error.
func MessageWrite(connection net.Conn, message []byte, obfuscated bool) (int, error) {
	if obfuscated {
		var err error
		message, err = obfuscate(message)
		if err != nil {
			return 0, err
		}
	}

	n, err := connection.Write(message)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func obfuscate(message []byte) ([]byte, error) {
	obfuscated := new(bytes.Buffer)

	var key [4]byte
	n, err := rand.Read(key[:])
	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, errors.New("could not read enough bytes for the key")
	}

	// Write the key to the obfuscated message.
	n, err = obfuscated.Write(key[:])
	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, errors.New("could not write enough bytes for the obfuscated key")
	}

	bufferedKey := new(bytes.Buffer)
	n, err = bufferedKey.Write(key[:])
	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, errors.New("could not write enough bytes for the buffered key")
	}

	rotatedKey, err := rotateKey(bufferedKey)
	if err != nil {
		return nil, err
	}

	newKey := new(bytes.Buffer)
	for k, v := range message {
		b, err := rotatedKey.ReadByte()
		if err != nil {
			return nil, err
		}

		// We construct the new key while reading the previous one for XOR.
		err = newKey.WriteByte(b)
		if err != nil {
			return nil, err
		}

		err = obfuscated.WriteByte(b ^ v)
		if err != nil {
			return nil, err
		}

		// We reached the 4 byte boundary.
		// Now we must rotate the key.
		if k%4 == 3 {
			rotatedKey, err = rotateKey(newKey)
			if err != nil {
				return nil, err
			}
		}
	}

	return obfuscated.Bytes(), nil
}

func rotateKey(key *bytes.Buffer) (*bytes.Buffer, error) {
	// Convert it to endian integer.
	var bigKey uint32
	err := binary.Read(key, binary.LittleEndian, &bigKey)
	if err != nil {
		return nil, err
	}

	// Shift 31 bits to the right.
	bigKey = (bigKey >> 31) | (bigKey << (32 - 31))

	// Convert back to little-endian byte array.
	rotatedKey := new(bytes.Buffer)
	err = binary.Write(rotatedKey, binary.LittleEndian, bigKey)
	if err != nil {
		return nil, err
	}

	return rotatedKey, nil
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
