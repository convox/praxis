package rack_test

import (
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestQueueStoreFetch(t *testing.T) {
	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	attrs := map[string]string{
		"data": "foo",
	}

	if err := rack.QueueStore("app", "key", attrs); !assert.NoError(t, err) {
		assert.FailNow(t, "unable to store message")

	}

	fm, err := rack.QueueFetch("app", "key", types.QueueFetchOptions{})
	assert.NoError(t, err)
	assert.Equal(t, attrs, fm)
}
