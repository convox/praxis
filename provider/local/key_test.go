package local_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	local, err := testProvider()
	assert.NoError(t, err)

	local.AppCreate("foo")

	data := "this is data to be encrypted"

	enc, err := local.KeyEncrypt("foo", "bar", []byte(data))
	assert.NoError(t, err)

	assert.NotEqual(t, []byte(data), enc)

	dec, err := local.KeyDecrypt("foo", "bar", enc)
	assert.NoError(t, err)

	assert.Equal(t, data, string(dec))
}
