package rack_test

import (
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableCreate(t *testing.T) {
	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	err = rack.TableCreate("app", "table", types.TableCreateOptions{Indexes: []string{"foo", "bar"}})
	assert.NoError(t, err)

	tab, err := rack.TableGet("app", "table")
	assert.NoError(t, err)

	assert.Equal(t, "table", tab.Name)
	assert.Equal(t, []string{"foo", "bar"}, tab.Indexes)
}

func TestTableList(t *testing.T) {
	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	zero, err := rack.TableList("app")
	assert.NoError(t, err)
	assert.Empty(t, zero)

	expects := []struct {
		Name    string
		Indexes []string
	}{
		// server returns orderd list by table name
		{"items", []string{"bar"}},
		{"logics", []string{"foo", "name"}},
		{"users", []string{"foo", "bar"}},
	}

	for _, tab := range expects {
		err = rack.TableCreate("app", tab.Name, types.TableCreateOptions{Indexes: tab.Indexes})
		assert.NoError(t, err)
	}

	tables, err := rack.TableList("app")
	assert.NoError(t, err)

	for i := range tables {
		assert.Equal(t, expects[i].Name, tables[i].Name)
		assert.Equal(t, expects[i].Indexes, tables[i].Indexes)
	}
}

func TestTableStoreFetch(t *testing.T) {
	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	if err := rack.TableCreate("app", "table", types.TableCreateOptions{}); !assert.NoError(t, err) {
		assert.FailNow(t, "table create failed")
	}

	id1, err := rack.TableStore("app", "table", map[string]string{"baz": "this", "data": "123456789"})
	assert.NoError(t, err)
	assert.NotEqual(t, "", id1)

	row, err := rack.TableFetch("app", "table", id1, types.TableFetchOptions{})
	assert.NoError(t, err)

	assert.Equal(t, "this", row["baz"])
	assert.Equal(t, "123456789", row["data"])
}

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
	assert.Empty(t, removed)
}

func TestTableTruncate(t *testing.T) {
	rack, err := setup()
	assert.NoError(t, err)
	defer cleanup()

	if err := rack.TableCreate("app", "table", types.TableCreateOptions{Indexes: []string{"data"}}); !assert.NoError(t, err) {
		assert.FailNow(t, "table create failed")
	}

	_, err = rack.TableStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	_, err = rack.TableStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	_, err = rack.TableStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	items, err := rack.TableFetchBatch("app", "table", []string{"foo"}, types.TableFetchOptions{Index: "data"})
	assert.NoError(t, err)
	assert.Len(t, items, 3)

	err = rack.TableTruncate("app", "table")
	assert.NoError(t, err)

	zero, err := rack.TableFetchBatch("app", "table", []string{"foo"}, types.TableFetchOptions{Index: "data"})
	assert.NoError(t, err)
	assert.Empty(t, zero)
}
