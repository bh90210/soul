package soul

import (
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadIP(t *testing.T) {
	tests := map[string]struct {
		input uint32
		want  net.IP
	}{
		"7.91.205.21":     {input: 123456789, want: []byte{7, 91, 205, 21}},
		"255.255.255.255": {input: math.MaxUint32, want: []byte{255, 255, 255, 255}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := ReadIP(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
