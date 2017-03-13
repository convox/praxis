package local_test

import (
	"sort"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableCreateGet(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	err = p.TableCreate("app", "table", types.TableCreateOptions{})
	assert.NoError(t, err)

	tt, err := p.TableGet("app", "table")
	assert.NoError(t, err)

	assert.Equal(t, "table", tt.Name)
	assert.Len(t, tt.Indexes, 0)
}

func TestTableList(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	zero, err := p.TableList("app")
	assert.NoError(t, err)
	assert.Empty(t, zero)

	expects := []struct {
		Name    string
		Indexes []string
	}{
		// orderd list by table name
		{"apples", []string{"bar"}},
		{"cherries", []string{"foo", "name"}},
		{"ninjas", []string{"foo", "bar"}},
	}

	for _, tab := range expects {
		err = p.TableCreate("app", tab.Name, types.TableCreateOptions{Indexes: tab.Indexes})
		assert.NoError(t, err)
	}

	tables, err := p.TableList("app")
	assert.NoError(t, err)

	for i := range tables {
		assert.Equal(t, expects[i].Name, tables[i].Name)
		assert.Equal(t, expects[i].Indexes, tables[i].Indexes)
	}
}

func TestTableTruncate(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	if err := p.TableCreate("app", "table", types.TableCreateOptions{Indexes: []string{"data"}}); !assert.NoError(t, err) {
		assert.FailNow(t, "table create failed")
	}

	_, err = p.TableRowStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	_, err = p.TableRowStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	_, err = p.TableRowStore("app", "table", map[string]string{"data": "foo"})
	assert.NoError(t, err)

	items, err := p.TableRowsGet("app", "table", []string{"foo"}, types.TableRowGetOptions{Index: "data"})
	assert.NoError(t, err)
	assert.Len(t, items, 3)

	err = p.TableTruncate("app", "table")
	assert.NoError(t, err)

	zero, err := p.TableRowsGet("app", "table", []string{"foo"}, types.TableRowGetOptions{Index: "data"})
	assert.NoError(t, err)
	assert.Empty(t, zero)
}

func TestTableRowGet(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	err = p.TableCreate("app", "table", types.TableCreateOptions{Indexes: []string{"foo"}})
	assert.NoError(t, err)

	_, err = p.TableRowGet("app", "table", "foo", types.TableRowGetOptions{})
	assert.EqualError(t, err, "not found")

	_, err = p.TableRowStore("app", "table", map[string]string{"foo": "bar"})
	assert.NoError(t, err)

	row, err := p.TableRowGet("app", "table", "bar", types.TableRowGetOptions{Index: "foo"})
	assert.NoError(t, err)
	assert.Equal(t, "bar", (*row)["foo"])

	_, err = p.TableRowStore("app", "table", map[string]string{"foo": "bar"})
	assert.NoError(t, err)

	_, err = p.TableRowStore("app", "table", map[string]string{"foo": "bar"})
	assert.NoError(t, err)

	_, err = p.TableRowGet("app", "table", "bar", types.TableRowGetOptions{Index: "foo"})
	assert.EqualError(t, err, "multiple items found")
}

func TestTableRowStore(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	err = p.TableCreate("app", "table", types.TableCreateOptions{Indexes: []string{"status"}})
	assert.NoError(t, err)

	id1, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar", "status": "running", "test": "change"})
	assert.NoError(t, err)

	row1, err := p.TableRowGet("app", "table", id1, types.TableRowGetOptions{})
	assert.NoError(t, err)

	assert.Equal(t, "bar", (*row1)["foo"])
	assert.Equal(t, "running", (*row1)["status"])
	assert.Equal(t, "change", (*row1)["test"])

	id2, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar", "status": "something", "id": id1})
	assert.NoError(t, err)

	if !assert.Equal(t, id1, id2) {
		assert.FailNow(t, "row IDs are not equal")
	}

	row2, err := p.TableRowGet("app", "table", "something", types.TableRowGetOptions{Index: "status"})
	assert.NoError(t, err)

	assert.Equal(t, "bar", (*row2)["foo"])
	assert.Equal(t, "something", (*row2)["status"])
	assert.Equal(t, "change", (*row2)["test"])
}

func TestTableRowStoreBytes(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	almostUTF8 := []byte{0x47, 0x51, 0xaf, 0x38, 0xdb, 0x23, 0x1, 0x5c, 0xc1, 0xa8, 0xc, 0x34, 0xcc, 0xc1, 0xef, 0x5c, 0x57, 0xa6, 0x92, 0x8, 0xe7, 0x6c, 0xcc, 0xfe, 0x1e, 0x1, 0x3, 0xe0, 0xed, 0xb2, 0x31, 0xdc, 0x2d, 0x37, 0x35, 0x36, 0x2c, 0x10, 0xa0, 0x6e, 0xf6, 0x39, 0xca, 0xb3, 0xbe, 0x3a, 0x99, 0xe1, 0x86, 0xbb, 0xaa, 0xca, 0x46, 0xd6, 0xe4, 0xf1}

	err = p.TableCreate("app", "table", types.TableCreateOptions{})
	assert.NoError(t, err)

	id, err := p.TableRowStore("app", "table", map[string]string{"foo": string(almostUTF8)})
	assert.NoError(t, err)

	row, err := p.TableRowGet("app", "table", id, types.TableRowGetOptions{})
	assert.NoError(t, err)

	assert.Equal(t, almostUTF8, []byte((*row)["foo"]))
}

func TestTableRowsGet(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	err = p.TableCreate("app", "table", types.TableCreateOptions{})
	assert.NoError(t, err)

	id1, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar1"})
	assert.NoError(t, err)

	id2, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar2"})
	assert.NoError(t, err)

	id3, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar3"})
	assert.NoError(t, err)

	items, err := p.TableRowsGet("app", "table", []string{id1, id2, id3}, types.TableRowGetOptions{})
	assert.NoError(t, err)

	sort.Slice(items, func(i, j int) bool { return items[i]["foo"] < items[j]["foo"] })

	if assert.Len(t, items, 3) {
		assert.Equal(t, "bar1", items[0]["foo"])
		assert.Equal(t, "bar2", items[1]["foo"])
		assert.Equal(t, "bar3", items[2]["foo"])
	}
}

func TestTableRowsDelete(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	opts := types.TableCreateOptions{
		Indexes: []string{"city"},
	}
	err = p.TableCreate("app", "table", opts)
	assert.NoError(t, err)

	id1, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar1", "city": "ATL"})
	assert.NoError(t, err)

	id2, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar2", "city": "MIA"})
	assert.NoError(t, err)

	id3, err := p.TableRowStore("app", "table", map[string]string{"foo": "bar3", "city": "ATL"})
	assert.NoError(t, err)

	items, err := p.TableRowsGet("app", "table", []string{id1, id2, id3}, types.TableRowGetOptions{})
	assert.NoError(t, err)

	sort.Slice(items, func(i, j int) bool { return items[i]["foo"] < items[j]["foo"] })

	if assert.Len(t, items, 3) {
		assert.Equal(t, "bar1", items[0]["foo"])
		assert.Equal(t, "bar2", items[1]["foo"])
		assert.Equal(t, "bar3", items[2]["foo"])
	}

	err = p.TableRowDelete("app", "table", id3, types.TableRowDeleteOptions{})
	assert.NoError(t, err)

	removed, err := p.TableRowsGet("app", "table", []string{id1, id2, id3}, types.TableRowGetOptions{})
	assert.NoError(t, err)

	sort.Slice(items, func(i, j int) bool { return items[i]["foo"] < items[j]["foo"] })

	if assert.Len(t, removed, 2) {
		assert.Equal(t, "bar1", removed[0]["foo"])
		assert.Equal(t, "bar2", removed[1]["foo"])
	}

	err = p.TableRowsDelete("app", "table", []string{"MIA", "ATL"}, types.TableRowDeleteOptions{Index: "city"})
	assert.NoError(t, err)

	none, err := p.TableRowsGet("app", "table", []string{id1, id2, id3}, types.TableRowGetOptions{})
	assert.NoError(t, err)
	assert.Len(t, none, 0)
}
