package soul

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendMessage(t *testing.T) {
	tests := map[string]struct {
		message    []byte
		messageLen int
		want       string
		error      bool
	}{
		"hello world": {
			message:    []byte{7, 91, 205, 21},
			want:       "hello world",
			messageLen: 4,
		},
		"error": {
			message: []byte{0},
			error:   true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server, client := net.Pipe()

			if tc.error {
				server.Close()
				_, err := MessageWrite(client, tc.message)
				assert.Error(t, err)

			} else {
				go func() {
					buf := new(bytes.Buffer)
					n, err := buf.ReadFrom(server)
					assert.NoError(t, err)
					assert.Equal(t, tc.messageLen, n)
					assert.Equal(t, tc.message, buf.Bytes())

					server.Close()
				}()

				i, err := MessageWrite(client, tc.message)
				assert.NoError(t, err)
				assert.Equal(t, tc.messageLen, i)
			}

			client.Close()
		})
	}
}

func TestReadMessage(t *testing.T) {
	tests := map[string]struct {
		want struct {
			reader io.Reader
			size   int
			code   ServerCode
		}
		error bool
	}{}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fmt.Println(tc)
		})
	}
}
