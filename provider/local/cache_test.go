package local_test

import (
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheStoreFetch(t *testing.T) {
	local, err := testProvider()
	require.NoError(t, err)
	defer testProviderCleanup(local)

	attrs := map[string]string{
		"test": "fake",
		"foo":  "bar",
	}

	err = local.CacheStore("app", "test", "data", attrs, types.CacheStoreOptions{})
	require.NoError(t, err)

	cached := local.CacheFetch("app", "test", "data")

	assert.Equal(t, "fake", cached["test"])
	assert.Equal(t, "bar", cached["foo"])
}
