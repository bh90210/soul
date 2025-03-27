//go:build testdata

package testdata

import "embed"

// Testdata contains an mp3 file for testing purposes.
//
//go:embed file_example_MP3_700KB.mp3
var Testdata embed.FS
