// Package peer messages are sent to peers over a P connection (TCP).
// Only a single active connection to a peer is allowed.
package peer

import "errors"

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

var ErrBanned = errors.New("Banned")

var ErrCancelled = errors.New("Cancelled")

var ErrComplete = errors.New("Complete")

var ErrFileNotShared = errors.New("File not shared.")

var ErrFileReadError = errors.New("File read error.")

var ErrPendingShutdown = errors.New("Pending shutdown.")

var ErrQueued = errors.New("Queued")

var ErrTooManyFiles = errors.New("Too many files")

var ErrTooManyMegabytes = errors.New("Too many megabytes")
