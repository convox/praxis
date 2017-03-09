package rack_test

import (
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableRemoveBatch(t *testing.T) {
	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	if err := rack.TableCreate("app", "table", types.TableCreateOptions{}); !assert.NoError(t, err) {
		assert.FailNow(t, "table create failed")
	}

	id1, err := rack.TableStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	id2, err := rack.TableStore("app", "table", map[string]string{"data": "bar"})
	assert.NoError(t, err)

	id3, err := rack.TableStore("app", "table", map[string]string{"data": "baz"})
	assert.NoError(t, err)

	items, err := rack.TableFetchBatch("app", "table", []string{id1, id2, id3}, types.TableFetchOptions{})
	assert.NoError(t, err)

	if assert.Len(t, items, 3) {
		assert.Equal(t, "foo", items[0]["data"])
		assert.Equal(t, "bar", items[1]["data"])
		assert.Equal(t, "baz", items[2]["data"])
	}

	err = rack.TableRemove("app", "table", id3, types.TableRemoveOptions{})
	assert.NoError(t, err)

	err = rack.TableRemoveBatch("app", "table", []string{id1, id2}, types.TableRemoveOptions{})
	assert.NoError(t, err)

	removed, err := rack.TableFetchBatch("app", "table", []string{id1, id2, id3}, types.TableFetchOptions{})
	assert.NoError(t, err)
	assert.Len(t, removed, 0)
}
