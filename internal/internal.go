package internal

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	"github.com/bh90210/soul"
)

type Code interface {
	soul.ServerCode | soul.PeerInitCode | soul.PeerCode | soul.DistributedCode
}

func MessageRead[C Code](c C, connection net.Conn) (io.Reader, int, C, error) {
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
	case soul.PeerInitCode, soul.DistributedCode:
		c, err := ReadUint8(messageHeader)
		if err != nil {
			return nil, 0, 0, err
		}

		code = C(c)

		readAlready = 1

	case soul.ServerCode, soul.PeerCode:
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

func MessageWrite(connection net.Conn, message []byte) (int, error) {
	n, err := connection.Write(message)
	if err != nil {
		return 0, err
	}

	return n, nil
}

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

func ReadInt32ToInt(buf io.Reader) (int, error) {
	val, err := ReadInt32(buf)
	if err != nil {
		return 0, err
	}

	return int(val), nil
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

func ReadUint32ToInt(buf io.Reader) (int, error) {
	v, err := ReadUint32(buf)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func ReadUint32ToToken(buf io.Reader) (soul.Token, error) {
	v, err := ReadUint32(buf)
	if err != nil {
		return 0, err
	}

	return soul.Token(v), nil
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

func ReadUint64ToInt(buf io.Reader) (int, error) {
	val, err := ReadUint64(buf)
	if err != nil {
		return 0, err
	}

	return int(val), nil
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
	binary.BigEndian.PutUint32(ip, val) // TODO: check why endianess is different than the rest.

	return ip
}
