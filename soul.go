// Package soul holds the common types and functions used by the rest of the packages.
package soul

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

// ConnectionType represents the type of connection.
type ConnectionType string

const (
	// Peer connection type: Peer To Peer.
	Peer ConnectionType = "P"
	// File connection type: File Transfer.
	File ConnectionType = "F"
	// Distributed connection type: Distributed Network.
	Distributed ConnectionType = "D"
)

var ErrMismatchingCodes = errors.New("mismatching codes")

const (
	MajorVersion uint32 = 1
	MinorVersion uint32 = 0
)

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

func ReadUint8(buf io.Reader) (uint8, error) {
	var val uint8
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func WriteUint8(buf io.Writer, val uint8) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func ReadInt32ToInt(buf io.Reader) (int, error) {
	val, err := ReadInt32(buf)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func ReadInt32(buf io.Reader) (int32, error) {
	var val int32
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func WriteInt32(buf io.Writer, val int32) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func ReadUint32ToInt(buf io.Reader) (int, error) {
	v, err := ReadUint32(buf)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func ReadUint32(buf io.Reader) (uint32, error) {
	var val uint32
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func WriteUint32(buf io.Writer, val uint32) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func ReadUint64ToInt(buf io.Reader) (int, error) {
	val, err := ReadUint64(buf)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func ReadInt64ToInt(buf io.Reader) (int, error) {
	val, err := ReadInt64(buf)
	if err != nil {
		return 0, err
	}

	return int(val), nil
}

func ReadInt64(buf io.Reader) (int64, error) {
	var val int64
	err := binary.Read(buf, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func WriteInt64(buf io.Writer, val int64) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func ReadUint64(reader io.Reader) (uint64, error) {
	var val uint64
	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func WriteUint64(buf io.Writer, val uint64) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func NewString(val string) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.WriteString(val)
	if err != nil {
		return nil, err
	}

	return Pack(buf.Bytes())
}

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

func ReadBool(reader io.Reader) (bool, error) {
	var val uint8

	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return false, err
	}

	return val == 1, nil
}

func WriteBool(buf io.Writer, val bool) error {
	var b uint8
	if val {
		b = 1
	}

	return binary.Write(buf, binary.LittleEndian, b)
}

func ReadBytes(reader io.Reader) ([]byte, error) {
	size, err := ReadUint32(reader)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, size)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

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

func ReadIP(val uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, val)

	return ip
}
