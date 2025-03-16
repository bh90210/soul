// Package peer messages are sent to peers over a P connection (TCP).
// Only a single active connection to a peer is allowed.
package peer

//go:generate stringer -type CodeInit -trimprefix Code
//go:generate stringer -type Code -trimprefix Code
//go:generate stringer -type UploadPermission
//go:generate stringer -type FileAttributeType
//go:generate stringer -type TransferDirection

import (
	"errors"
	"io"
	"net"

	"github.com/bh90210/soul"
	"github.com/bh90210/soul/internal"
)

// ConnectionType represents the type of peer 'P' connection.
const ConnectionType soul.ConnectionType = "P"

// CodeInit Peer init messages are used to initiate a P, F or D connection (TCP) to a peer.
type CodeInit int

// Code Peer messages are sent to peers over a P connection (TCP).
// Only a single active connection to a peer is allowed.
type Code int

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

// TransferDirection represents the direction of a file transfer.
type TransferDirection int

const (
	// DownloadFromPeer transfer direction.
	DownloadFromPeer TransferDirection = iota
	// UploadToPeer transfer direction.
	UploadToPeer
)

// ErrBanned is returned when the transfer is not allowed because the peer is banned.
var ErrBanned = errors.New("Banned")

// ErrCancelled is returned when the transfer is not allowed because it was cancelled.
var ErrCancelled = errors.New("Cancelled")

// ErrComplete is returned when the transfer is complete.
var ErrComplete = errors.New("Complete")

// ErrFileNotShared is returned when the transfer is not allowed because the file is not shared.
var ErrFileNotShared = errors.New("File not shared.")

// ErrFileReadError is returned when the transfer is not allowed because of a file read error.
var ErrFileReadError = errors.New("File read error.")

// ErrPendingShutdown is returned when the transfer is not allowed because of a pending shutdown.
var ErrPendingShutdown = errors.New("Pending shutdown.")

// ErrQueued is returned when the transfer is queued.
var ErrQueued = errors.New("Queued")

// ErrTooManyFiles is returned when the transfer is not allowed because there are too many files.
var ErrTooManyFiles = errors.New("Too many files")

// ErrTooManyMegabytes is returned when the transfer is not allowed because there are too many megabytes.
var ErrTooManyMegabytes = errors.New("Too many megabytes")

func Reason(reason string) error {
	switch reason {
	case ErrBanned.Error():
		return ErrBanned
	case ErrCancelled.Error():
		return ErrCancelled
	case ErrComplete.Error():
		return ErrComplete
	case ErrFileNotShared.Error():
		return ErrFileNotShared
	case ErrFileReadError.Error():
		return ErrFileReadError
	case ErrPendingShutdown.Error():
		return ErrPendingShutdown
	case ErrQueued.Error():
		return ErrQueued
	case ErrTooManyFiles.Error():
		return ErrTooManyFiles
	case ErrTooManyMegabytes.Error():
		return ErrTooManyMegabytes

	default:
		return errors.New(reason)
	}
}

func Read[C CodeInit | Code](c C, connection net.Conn, obfuscated bool) (io.Reader, int, C, error) {
	switch any(c).(type) {
	case CodeInit:
		r, s, code, err := internal.MessageRead(internal.CodePeerInit(0), connection, obfuscated)
		return r, s, C(code), err

	case Code:
		r, s, code, err := internal.MessageRead(internal.CodePeer(0), connection, obfuscated)
		return r, s, C(code), err

	default:
		return nil, 0, 0, errors.New("invalid code")
	}
}

type message[M any] interface {
	*PeerInit |
		*PierceFirewall |
		*FileSearchResponse |
		*FolderContentsRequest |
		*FolderContentsResponse |
		*PlaceInQueueRequest |
		*PlaceInQueueResponse |
		*QueueUpload |
		*SharedFileListRequest |
		*SharedFileListResponse |
		*TransferRequest |
		*TransferResponse |
		*UploadDenied |
		*UploadFailed |
		*UserInfoRequest |
		*UserInfoResponse
	Serialize(M) ([]byte, error)
}

func Write[M message[M]](connection net.Conn, message M, obfuscate bool) (int, error) {
	m, err := message.Serialize(message)
	if err != nil {
		return 0, err
	}

	return internal.MessageWrite(connection, m, obfuscate)
}
