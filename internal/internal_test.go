package internal

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageRead(t *testing.T) {
	t.Parallel()

	t.Run("Uint8", func(t *testing.T) {
		buf := new(bytes.Buffer)

		err := WriteUint8(buf, uint8(0)) // Code.
		assert.NoError(t, err)

		err = WriteUint32(buf, 1) // Data.
		assert.NoError(t, err)

		b, err := Pack(buf.Bytes())
		assert.NoError(t, err)

		r, size, code, err := MessageRead(CodePeerInit(0), bytes.NewBuffer(b), false)
		assert.NoError(t, err)
		assert.Equal(t, uint32(5), size)
		assert.Equal(t, CodePeerInit(0), code)

		s, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, uint32(5), s)

		c, err := ReadUint8(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint8(0), c)

		m, err := ReadUint32(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), m)

	})

	t.Run("Uint32", func(t *testing.T) {
		buf := new(bytes.Buffer)

		err := WriteUint32(buf, 2) // Code.
		assert.NoError(t, err)

		err = WriteUint32(buf, 1) // Data.
		assert.NoError(t, err)

		b, err := Pack(buf.Bytes())
		assert.NoError(t, err)

		r, size, code, err := MessageRead(CodeServer(0), bytes.NewBuffer(b), false)
		assert.NoError(t, err)
		assert.Equal(t, uint32(8), size)
		assert.Equal(t, CodeServer(2), code)

		s, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, uint32(8), s)

		c, err := ReadUint32(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint32(2), c)

		m, err := ReadUint32(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), m)
	})

	t.Run("Obfuscated Uint32", func(t *testing.T) {
		message := new(bytes.Buffer)

		err := WriteUint32(message, 1) // Code.
		assert.NoError(t, err)

		token := soul.NewToken()
		err = WriteUint32(message, uint32(token)) // Data.
		assert.NoError(t, err)

		b, err := Pack(message.Bytes())
		assert.NoError(t, err)

		message.Reset()
		n, err := MessageWrite(message, b, true)
		assert.NoError(t, err)
		assert.Equal(t, 16, n)

		r, size, code, err := MessageRead(CodePeer(0), message, true)
		assert.NoError(t, err)
		require.Equal(t, uint32(8), size)
		assert.Equal(t, CodePeer(1), code)
		require.NotNil(t, r)

		s, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, uint32(8), s)

		c, err := ReadUint32(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), c)

		m, err := ReadUint32(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, uint32(token), m)
	})
}

func TestMessageWrite(t *testing.T) {
	t.Parallel()

	message := new(bytes.Buffer)
	err := WriteUint32(message, 1)
	assert.NoError(t, err)

	err = WriteString(message, "test")
	assert.NoError(t, err)

	packed, err := Pack(message.Bytes())
	assert.NoError(t, err)

	t.Run("Non obfuscated", func(t *testing.T) {
		buf := new(bytes.Buffer)
		n, err := MessageWrite(buf, packed, false)
		assert.NoError(t, err)
		assert.Equal(t, 16, n)

		r, s, c, err := MessageRead(CodePeer(0), buf, false)
		assert.NoError(t, err)
		assert.Equal(t, 12, int(s))
		assert.Equal(t, CodePeer(1), c)

		size, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, 12, int(size))

		code, err := ReadUint32(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), code)

		data, err := ReadString(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, "test", data)
	})

	t.Run("Obfuscated", func(t *testing.T) {
		buf := new(bytes.Buffer)
		n, err := MessageWrite(buf, packed, true)
		assert.NoError(t, err)
		assert.Equal(t, 20, n)

		r, s, c, err := MessageRead(CodePeer(0), buf, true)
		assert.NoError(t, err)
		assert.Equal(t, 12, int(s))
		assert.Equal(t, CodePeer(1), c)

		size, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, 12, int(size))

		code, err := ReadUint32(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint32(1), code)

		data, err := ReadString(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, "test", data)
	})

	t.Run("Init Obfuscated", func(t *testing.T) {
		message := new(bytes.Buffer)
		err := WriteUint8(message, 1)
		assert.NoError(t, err)

		err = WriteString(message, "test")
		assert.NoError(t, err)

		packed, err := Pack(message.Bytes())
		assert.NoError(t, err)

		buf := new(bytes.Buffer)
		n, err := MessageWrite(buf, packed, true)
		assert.NoError(t, err)
		assert.Equal(t, 17, n)
		assert.Equal(t, 17, buf.Len())

		r, s, c, err := MessageRead(CodePeerInit(0), buf, true)
		assert.NoError(t, err)
		assert.Equal(t, 9, int(s))
		assert.Equal(t, CodePeerInit(1), c)

		size, err := ReadUint32(r) // Size.
		assert.NoError(t, err)
		assert.Equal(t, 9, int(size))

		code, err := ReadUint8(r) // Code.
		assert.NoError(t, err)
		assert.Equal(t, uint8(1), code)

		data, err := ReadString(r) // Data.
		assert.NoError(t, err)
		assert.Equal(t, "test", data)
	})
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

	t.Run("EOF", func(t *testing.T) {
		buf := new(bytes.Buffer)
		_, err := ReadUint8(buf)
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("Success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err := buf.WriteByte(1)
		assert.NoError(t, err)
		expected := uint8(1)

		actual, err := ReadUint8(buf)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

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
