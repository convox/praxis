package rack_test

import (
	"testing"

	"github.com/convox/praxis/cycle"
	"github.com/stretchr/testify/assert"
)

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
