package internal

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestMessageRead(t *testing.T) {
	t.Parallel()

	t.Run("Uint8", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		server, client := net.Pipe()
		go func() {
			defer wg.Done()

			buf := new(bytes.Buffer)

			err := WriteUint8(buf, uint8(0)) // Code.
			assert.NoError(t, err)

			err = WriteUint32(buf, 1) // Data.
			assert.NoError(t, err)

			b, err := Pack(buf.Bytes())
			assert.NoError(t, err)

			n, err := server.Write(b)
			assert.NoError(t, err)
			assert.Equal(t, 9, n)

		}()

		r, size, code, err := MessageRead(soul.CodePeerInit(0), client)
		assert.NoError(t, err)
		assert.Equal(t, 5, size)
		assert.Equal(t, soul.CodePeerInit(0), code)

		s, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, uint32(5), s)

		c, err := ReadUint8(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint8(0), c)

		m, err := ReadUint32(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), m)

		wg.Wait()
	})

	t.Run("Uint32", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		server, client := net.Pipe()
		go func() {
			defer wg.Done()

			buf := new(bytes.Buffer)

			err := WriteUint32(buf, 0) // Code.
			assert.NoError(t, err)

			err = WriteUint32(buf, 1) // Data.
			assert.NoError(t, err)

			b, err := Pack(buf.Bytes())
			assert.NoError(t, err)

			n, err := server.Write(b)
			assert.NoError(t, err)
			assert.Equal(t, 12, n)
		}()

		r, size, code, err := MessageRead(soul.CodeServer(0), client)
		assert.NoError(t, err)
		assert.Equal(t, 8, size)
		assert.Equal(t, soul.CodeServer(0), code)

		s, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, uint32(8), s)

		c, err := ReadUint32(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), c)

		m, err := ReadUint32(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), m)

		wg.Wait()
	})

}

func TestMessageWrite(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	wg.Add(1)

	server, client := net.Pipe()
	go func() {
		defer wg.Done()

		buf := make([]byte, 1)
		n, err := server.Read(buf)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, []byte{1}, buf)
	}()

	n, err := MessageWrite(client, []byte{1})
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	wg.Wait()
}
func TestPack(t *testing.T) {
	t.Parallel()

	expected := []byte{1, 0, 0, 0, 1}

	actual, err := Pack([]byte{1})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

}

func TestReadUint8(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := buf.WriteByte(1)
	assert.NoError(t, err)
	expected := uint8(1)

	actual, err := ReadUint8(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteUint8(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := uint8(1)

	err := WriteUint8(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{1}, actual)
}

func TestReadInt32(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, int32(1))
	assert.NoError(t, err)
	expected := int32(1)

	actual, err := ReadInt32(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteInt32(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := int32(1)

	err := WriteInt32(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{1, 0, 0, 0}, actual)
}

func TestReadInt32ToInt(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, int32(1))
	assert.NoError(t, err)
	expected := 1

	actual, err := ReadInt32ToInt(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func TestReadUint32(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(1))
	assert.NoError(t, err)
	expected := uint32(1)

	actual, err := ReadUint32(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteUint32(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := uint32(1)

	err := WriteUint32(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{1, 0, 0, 0}, actual)
}

func TestReadUint32ToInt(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(1))
	assert.NoError(t, err)
	expected := 1

	actual, err := ReadUint32ToInt(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReadUint32ToToken(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(1))
	assert.NoError(t, err)
	expected := soul.Token(1)

	actual, err := ReadUint32ToToken(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReadUint64(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint64(1))
	assert.NoError(t, err)
	expected := uint64(1)

	actual, err := ReadUint64(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteUint64(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := uint64(1)

	err := WriteUint64(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{1, 0, 0, 0, 0, 0, 0, 0}, actual)
}

func TestReadUint64ToInt(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint64(1))
	assert.NoError(t, err)
	expected := 1

	actual, err := ReadUint64ToInt(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestReadString(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(3))
	assert.NoError(t, err)
	err = binary.Write(buf, binary.LittleEndian, []byte("foo"))
	assert.NoError(t, err)
	expected := "foo"

	actual, err := ReadString(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteString(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := "foo"

	err := WriteString(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{3, 0, 0, 0, 102, 111, 111}, actual)
}

func TestReadBool(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := buf.WriteByte(1)
	assert.NoError(t, err)
	expected := true

	actual, err := ReadBool(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteBool(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := true

	err := WriteBool(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{1}, actual)
}

func TestReadBytes(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, uint32(3))
	assert.NoError(t, err)
	err = binary.Write(buf, binary.LittleEndian, []byte("foo"))
	assert.NoError(t, err)
	expected := []byte("foo")

	actual, err := ReadBytes(buf)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestWriteBytes(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	expected := []byte("foo")

	err := WriteBytes(buf, expected)
	assert.NoError(t, err)

	actual := buf.Bytes()
	assert.Equal(t, []byte{3, 0, 0, 0, 102, 111, 111}, actual)
}

func TestReadIP(t *testing.T) {
	t.Parallel()

	actual := ReadIP(1)
	assert.Equal(t, net.IP{0, 0, 0, 1}, actual)
}
