// Package file messages are sent to peers over a F connection (TCP),
// and do not have messages codes associated with them.
package file

import "github.com/bh90210/soul"

// ConnectionType represents the type of file 'F' connection.
const ConnectionType soul.ConnectionType = "F"
