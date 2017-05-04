package controllers_test

import (
	"io/ioutil"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableGet(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	table := &types.Table{
		Name:    "table",
		Indexes: []string{"foo", "baz"},
	}
	mp.On("TableGet", "app", "table").Return(table, nil)

	res, err := testRequest(ts, "GET", "/apps/app/tables/table", nil)
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(`{"Name":"table","Indexes":["foo","baz"]}`), data)
}

func TestTableList(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	tables := types.Tables{
		{Name: "table", Indexes: []string{"foo", "baz"}},
		{Name: "table2", Indexes: []string{"baz"}},
		{Name: "table1", Indexes: []string{"floor"}},
	}

	mp.On("TableList", "app").Return(tables, nil)

	res, err := testRequest(ts, "GET", "/apps/app/tables", nil)
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, string(`[{"Name":"table","Indexes":["foo","baz"]},{"Name":"table1","Indexes":["floor"]},{"Name":"table2","Indexes":["baz"]}]`), string(data))
}

func TestTableTruncate(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	mp.On("TableTruncate", "app", "table").Return(nil)

	res, err := testRequest(ts, "POST", "/apps/app/tables/table/truncate", nil)
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(""), data)
}
