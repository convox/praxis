package rack_test

import (
	"testing"

	"github.com/convox/praxis/cycle"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestQueueStore(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "POST", Path: "/apps/app/queues/queue", Body: []byte("data=foo")},
		cycle.HTTPResponse{Code: 200},
	)

	err := r.QueueStore("app", "queue", map[string]string{"data": "foo"})
	assert.NoError(t, err)
}

func TestQueueFetch(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "GET", Path: "/apps/app/queues/queue"},
		cycle.HTTPResponse{Code: 200, Body: []byte(`{"data":"foo"}`)},
	)

	attrs, err := r.QueueFetch("app", "queue", types.QueueFetchOptions{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"data": "foo"}, attrs)
}
