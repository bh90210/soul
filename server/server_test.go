package server

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	t.Parallel()

	expected := new(Login)
	expected.Username = "test"
	expected.Password = "test"
	message, err := expected.Serialize(expected)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	r, size, code, err := Read(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 64, size)
	assert.Equal(t, CodeLogin, code)
}

func TestWrite(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	i, err := Write(buf, &Login{Username: "test", Password: "test"})
	assert.NoError(t, err)
	assert.Equal(t, 68, i)

	r, size, code, err := Read(buf)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 64, size)
	assert.Equal(t, CodeLogin, code)
}
