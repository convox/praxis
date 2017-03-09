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
