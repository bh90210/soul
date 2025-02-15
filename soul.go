package soul

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
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
	// P connection type: Peer To Peer.
	P ConnectionType = "P"
	// F connection type: File Transfer.
	F ConnectionType = "F"
	// D connection type: Distributed Network.
	D ConnectionType = "D"
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

func Pack(data []byte) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, UInt(len(data)))
	binary.Write(buf, binary.LittleEndian, data)

	return buf.Bytes()
}

func ReadUInt(reader io.Reader) UInt {
	var val UInt
	binary.Read(reader, binary.LittleEndian, &val)

	return val
}

func WriteUInt(buf *bytes.Buffer, val UInt) {
	binary.Write(buf, binary.LittleEndian, val)
}

func NewString(content string) String {
	buf := new(bytes.Buffer)
	buf.WriteString(content)

	return Pack(buf.Bytes())
}

func ReadString(reader io.Reader) string {
	size := ReadUInt(reader)
	buf := make([]byte, size)
	io.ReadFull(reader, buf)

	return string(buf)
}

func ReadBool(reader io.Reader) bool {
	var val Boolean

	binary.Read(reader, binary.LittleEndian, &val)

	return val == 1
}

func ReadIP(val UInt) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(val))

	return ip
}

func Read(connection io.Reader) (io.Reader, UInt, UInt) {
	messageCopy := new(bytes.Buffer)
	message := io.TeeReader(connection, messageCopy)

	// Read the messageSize of the message.
	packetSize := ReadUInt(message)

	// Read the code.
	code := ReadUInt(message)

	// The size of the actual message needs -4 to account for the packetSize and code.
	messageSize := packetSize - 4

	sizeSoFar := 0
	for {
		p := make([]byte, int(messageSize)-sizeSoFar)
		n, err := message.Read(p)
		if err != nil {
			log.Fatal("Read error", err)
		}

		sizeSoFar += n

		if sizeSoFar == int(messageSize) {
			break
		}
	}

	return messageCopy, packetSize, code
}
