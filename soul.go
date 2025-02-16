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

// UserStatus represents the status of a user.
type UserStatus int

const (
	// Offline user status.
	Offline UserStatus = iota
	// Away user status.
	Away
	// Online user status.
	Online
)

// UploadPermission represents the permission level for uploading files.
type UploadPermission int

const (
	// NoOne permission level.
	NoOne UploadPermission = iota
	// Everyone permission level.
	Everyone
	// UsersInList permission level.
	UsersInList
	// PermittedUsers permission level.
	PermittedUsers
)

// TransferDirection represents the direction of a file transfer.
type TransferDirection int

const (
	// DownloadFromPeer transfer direction.
	DownloadFromPeer TransferDirection = iota
	// UploadToPeer transfer direction.
	UploadToPeer
)

var ErrMismatchingCodes = errors.New("mismatching codes")

var ErrTransferRejectionBanned = errors.New("Banned")

var ErrTransferRejectionCancelled = errors.New("Cancelled")

var ErrTransferRejectionComplete = errors.New("Complete")

var ErrTransferRejectionFileNotShared = errors.New("File not shared.")

var ErrTransferRejectionFileReadError = errors.New("File read error.")

var ErrTransferRejectionPendingShutdown = errors.New("Pending shutdown.")

var ErrTransferRejectionQueued = errors.New("Queued")

var ErrTransferRejectionTooManyFiles = errors.New("Too many files")

var ErrTransferRejectionTooManyMegabytes = errors.New("Too many megabytes")

// FileAttributeType represents the type of file attribute.
type FileAttributeType int

const (
	// Bitrate (kbps).
	Bitrate FileAttributeType = iota
	// Duration (seconds).
	Duration
	// VBR (0 or 1).
	VBR
	// Encoder (unused). See https://nicotine-plus.org/doc/SLSKPROTOCOL.html#file-attribute-types.
	_
	// SampleRate (Hz).
	SampleRate
	// BitDepth (bits).
	BitDepth
)

const (
	MajorVersion uint32 = 160
	MinorVersion uint32 = 1
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

func ReadUint32(reader io.Reader) (uint32, error) {
	var val uint32
	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func ReadInt(reader io.Reader) (int, error) {
	v, err := ReadUint32(reader)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func WriteUint32(buf *bytes.Buffer, val uint32) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func NewString(content string) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.WriteString(content)
	if err != nil {
		return nil, err
	}

	return Pack(buf.Bytes())
}

func WriteString(buf *bytes.Buffer, content string) error {
	c, err := NewString(content)
	if err != nil {
		return err
	}

	err = binary.Write(buf, binary.LittleEndian, c)
	if err != nil {
		return err
	}

	return nil
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

func ReadBool(reader io.Reader) (bool, error) {
	var val uint8

	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return false, err
	}

	return val == 1, nil
}

func WriteBool(buf *bytes.Buffer, val bool) error {
	var b uint8
	if val {
		b = 1
	}

	return binary.Write(buf, binary.LittleEndian, b)
}

func ReadIP(val uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, val)

	return ip
}
