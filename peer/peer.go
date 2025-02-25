// Package peer messages are sent to peers over a P connection (TCP).
// Only a single active connection to a peer is allowed.
package peer

//go:generate stringer -type CodeInit,Code -trimprefix Code
//go:generate stringer -type UploadPermission -trimprefix Permission
//go:generate stringer -type FileAttributeType -trimprefix Attribute
//go:generate stringer -type TransferDirection -trimprefix Direction

import (
	"errors"

	"github.com/bh90210/soul"
)

// ConnectionType represents the type of peer 'P' connection.
const ConnectionType soul.ConnectionType = "P"

// CodeInit Peer init messages are used to initiate a P, F or D connection (TCP) to a peer.
type CodeInit soul.CodePeerInit

const (
	CodePierceFirewall CodeInit = 0
	CodePeerInit
)

// Code Peer messages are sent to peers over a P connection (TCP).
// Only a single active connection to a peer is allowed.
type Code soul.CodePeer

const (
	CodeSharedFileListResponse Code = 5
)

// UploadPermission represents the permission level for uploading files.
type UploadPermission int

const (
	// PermissionNoOne permission level.
	PermissionNoOne UploadPermission = iota
	// PermissionEveryone permission level.
	PermissionEveryone
	// PermissionUsersInList permission level.
	PermissionUsersInList
	// PermissionPermittedUsers permission level.
	PermissionPermittedUsers
)

// FileAttributeType represents the type of file attribute.
type FileAttributeType int

const (
	// AttributeBitrate (kbps).
	AttributeBitrate FileAttributeType = iota
	// AttributeDuration (seconds).
	AttributeDuration
	// AttributeVBR (0 or 1).
	AttributeVBR
	// Encoder (unused). See https://nicotine-plus.org/doc/SLSKPROTOCOL.html#file-attribute-types.
	_
	// AttributeSampleRate (Hz).
	AttributeSampleRate
	// AttributeBitDepth (bits).
	AttributeBitDepth
)

// TransferDirection represents the direction of a file transfer.
type TransferDirection int

const (
	// DirectionDownloadFromPeer transfer direction.
	DirectionDownloadFromPeer TransferDirection = iota
	// DirectionUploadToPeer transfer direction.
	DirectionUploadToPeer
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
