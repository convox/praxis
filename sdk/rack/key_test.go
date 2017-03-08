package rack_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	plainText = "this is data to be encrypted"
	encHex    = "0109a2789d79141f96fd1d37dd8cb94d00f1d35ced3c87be0c4305b64aaabe40ccff1a580dacd69ce6d842fd"
)

func TestEncrypt(t *testing.T) {

	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	enc, err := rack.KeyEncrypt("app", "key", []byte(plainText))
	assert.NoError(t, err)
	assert.Equal(t, encHex, fmt.Sprintf("%x", enc))
}

func TestDecrypt(t *testing.T) {

	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	bytes, err := hex.DecodeString(encHex)
	assert.NoError(t, err)

	dec, err := rack.KeyDecrypt("app", "key", []byte(bytes))
	assert.NoError(t, err)
	assert.Equal(t, plainText, string(dec))
}
