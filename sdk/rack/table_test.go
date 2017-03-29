package rack_test

import (
	"testing"

	"github.com/convox/praxis/cycle"
	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableCreate(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "POST", Path: "/apps/app/tables/table", Body: []byte("")},
		cycle.HTTPResponse{Code: 200},
	)

	err := r.TableCreate("app", "table", types.TableCreateOptions{})
	assert.NoError(t, err)
}

func TestTableCreateIndexes(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "POST", Path: "/apps/app/tables/table", Body: []byte("index=foo&index=bar")},
		cycle.HTTPResponse{Code: 200},
	)

	err := r.TableCreate("app", "table", types.TableCreateOptions{Indexes: []string{"foo", "bar"}})
	assert.NoError(t, err)
}

func TestTableList(t *testing.T) {
	r, c := testRack()

	c.Add(
		cycle.HTTPRequest{Method: "GET", Path: "/apps/app/tables", Body: []byte("")},
		cycle.HTTPResponse{Code: 200, Body: []byte(`[{"name":"t1","indexes":["foo","bar"]}]`)},
	)

	tables, err := r.TableList("app")
	assert.NoError(t, err)

	if assert.Len(t, tables, 1) {
		assert.Equal(t, "t1", tables[0].Name)
		assert.Equal(t, []string{"foo", "bar"}, tables[0].Indexes)
	}
}
