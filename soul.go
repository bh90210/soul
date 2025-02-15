package soul

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

type (
	//Int represents a signed 32-bit integer.
	Int int32
	// UInt represents an unsigned 32-bit integer.
	UInt uint32
	// Long represents a signed 64-bit integer.
	Long int64
	// ULong represents an unsigned 64-bit integer.
	ULong uint64
	// String represents a string.
	String []byte
	// Boolean represents a boolean.
	Boolean uint8
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

// UserStatusCode represents the status of a user.
type UserStatusCode int

const (
	// Offline user status.
	Offline UserStatusCode = iota
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
	MajorVersion UInt = 160
	MinorVersion UInt = 1
)

func Pack(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, UInt(len(data)))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func ReadUInt(reader io.Reader) (UInt, error) {
	var val UInt
	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func ReadInt(reader io.Reader) (int, error) {
	v, err := ReadUInt(reader)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func WriteUInt(buf *bytes.Buffer, val UInt) error {
	return binary.Write(buf, binary.LittleEndian, val)
}

func NewString(content string) (String, error) {
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
	size, err := ReadUInt(reader)
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
	var val Boolean

	err := binary.Read(reader, binary.LittleEndian, &val)
	if err != nil {
		return false, err
	}

	return val == 1, nil
}

func WriteBool(buf *bytes.Buffer, val bool) error {
	var b Boolean
	if val {
		b = 1
	}

	return binary.Write(buf, binary.LittleEndian, b)
}

func ReadIP(val UInt) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(val))

	return ip
}
