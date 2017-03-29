package rack_test

import (
	"testing"

	"github.com/alecthomas/assert"
	"github.com/convox/praxis/cycle"
)

func TestKeyDecrypt(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "POST", Path: "/apps/app/keys/key/decrypt", Body: []byte("eNcRyPtEd")},
		cycle.HTTPResponse{Code: 200, Body: []byte("message")},
	)

	data, err := r.KeyDecrypt("app", "key", []byte("eNcRyPtEd"))
	assert.NoError(t, err)
	assert.Equal(t, "message", string(data))
}

func TestKeyEncrypt(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "POST", Path: "/apps/app/keys/key/encrypt", Body: []byte("message")},
		cycle.HTTPResponse{Code: 200, Body: []byte("eNcRyPtEd")},
	)

	data, err := r.KeyEncrypt("app", "key", []byte("message"))
	assert.NoError(t, err)
	assert.Equal(t, "eNcRyPtEd", string(data))
}
