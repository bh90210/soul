package peer

import (
	"bytes"
	"testing"

	"github.com/bh90210/soul"
	"github.com/stretchr/testify/assert"
)

func TestPierceFirewall(t *testing.T) {
	t.Parallel()

	pf := new(PierceFirewall)
	pf.Token = soul.NewToken()
	message, err := pf.Serialize(pf)
	assert.NoError(t, err)
	assert.NotNil(t, message)

	des := new(PierceFirewall)
	err = des.Deserialize(bytes.NewReader(message))
	assert.NoError(t, err)
	assert.Equal(t, pf, des)
}
