package local_test

import (
	"sort"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableFetchBatch(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	err = p.TableCreate("app", "table", types.TableCreateOptions{})
	assert.NoError(t, err)

	id1, err := p.TableStore("app", "table", map[string]string{"foo": "bar1"})
	assert.NoError(t, err)

	id2, err := p.TableStore("app", "table", map[string]string{"foo": "bar2"})
	assert.NoError(t, err)

	id3, err := p.TableStore("app", "table", map[string]string{"foo": "bar3"})
	assert.NoError(t, err)

	items, err := p.TableFetchBatch("app", "table", []string{id1, id2, id3}, types.TableFetchOptions{})
	assert.NoError(t, err)

	sort.Slice(items, func(i, j int) bool { return items[i]["foo"] < items[j]["foo"] })

	if assert.Len(t, items, 3) {
		assert.Equal(t, "bar1", items[0]["foo"])
		assert.Equal(t, "bar2", items[1]["foo"])
		assert.Equal(t, "bar3", items[2]["foo"])
	}
}

func TestTableRemoveBatch(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	opts := types.TableCreateOptions{
		Indexes: []string{"city"},
	}
	err = p.TableCreate("app", "table", opts)
	assert.NoError(t, err)

	id1, err := p.TableStore("app", "table", map[string]string{"foo": "bar1", "city": "ATL"})
	assert.NoError(t, err)

	id2, err := p.TableStore("app", "table", map[string]string{"foo": "bar2", "city": "MIA"})
	assert.NoError(t, err)

	id3, err := p.TableStore("app", "table", map[string]string{"foo": "bar3", "city": "ATL"})
	assert.NoError(t, err)

	items, err := p.TableFetchBatch("app", "table", []string{id1, id2, id3}, types.TableFetchOptions{})
	assert.NoError(t, err)

	sort.Slice(items, func(i, j int) bool { return items[i]["foo"] < items[j]["foo"] })

	if assert.Len(t, items, 3) {
		assert.Equal(t, "bar1", items[0]["foo"])
		assert.Equal(t, "bar2", items[1]["foo"])
		assert.Equal(t, "bar3", items[2]["foo"])
	}

	err = p.TableRemove("app", "table", id3, types.TableRemoveOptions{})
	assert.NoError(t, err)

	removed, err := p.TableFetchBatch("app", "table", []string{id1, id2, id3}, types.TableFetchOptions{})
	assert.NoError(t, err)

	sort.Slice(items, func(i, j int) bool { return items[i]["foo"] < items[j]["foo"] })

	if assert.Len(t, removed, 2) {
		assert.Equal(t, "bar1", removed[0]["foo"])
		assert.Equal(t, "bar2", removed[1]["foo"])
	}

	err = p.TableRemoveBatch("app", "table", []string{"MIA", "ATL"}, types.TableRemoveOptions{Index: "city"})
	assert.NoError(t, err)

	none, err := p.TableFetchBatch("app", "table", []string{id1, id2, id3}, types.TableFetchOptions{})
	assert.NoError(t, err)
	assert.Len(t, none, 0)
}
