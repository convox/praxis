package controllers_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableCreate(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	mp.On("TableCreate", "app", "table1", types.TableCreateOptions{Indexes: []string{"floor"}}).Return(nil)

	v := url.Values{}
	v.Add("index", "floor")

	res, err := testRequest(ts, "POST", "/apps/app/tables/table1", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(``), data)

	mp.On("TableCreate", "app", "tableerror", types.TableCreateOptions{Indexes: []string{"floor"}}).Return(fmt.Errorf("failed to create table"))
	reserr, err := testRequest(ts, "POST", "/apps/app/tables/tableerror", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer reserr.Body.Close()

	data, err = ioutil.ReadAll(reserr.Body)
	assert.NoError(t, err)

	assert.Equal(t, 500, reserr.StatusCode)
	assert.Equal(t, []byte("failed to create table\n"), data)
}

func TestTableRowGet(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	row := types.TableRow{
		"foo":  "bar",
		"test": "test",
		"id":   "idfoo",
	}

	mp.On("TableRowGet", "app", "table", "idfoo", types.TableRowGetOptions{Index: "id"}).Return(&row, nil)

	res, err := testRequest(ts, "GET", "/apps/app/tables/table/indexes/id/idfoo", nil)
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(`{"foo":"bar","id":"idfoo","test":"test"}`), data)
}

func TestTableRowsGet(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	items := types.TableRows{
		types.TableRow{"id": "one", "foo": "bar1"},
		types.TableRow{"id": "two", "foo": "bar2"},
	}

	mp.On("TableRowsGet", "app", "table", []string{"one", "two"}, types.TableRowGetOptions{Index: "id"}).Return(items, nil)

	v := url.Values{}
	v.Add("key", "one")
	v.Add("key", "two")

	res, err := testRequest(ts, "POST", "/apps/app/tables/table/indexes/id/batch", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(`[{"foo":"bar1","id":"one"},{"foo":"bar2","id":"two"}]`), data)
}

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
	assert.Equal(t, []byte(`[{"Name":"table","Indexes":["foo","baz"]},{"Name":"table2","Indexes":["baz"]},{"Name":"table1","Indexes":["floor"]}]`), data)
}

func TestTableRowStore(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	attrs := types.TableRow{
		"foo":  "bar",
		"test": "this",
	}

	mp.On("TableRowStore", "app", "table", attrs).Return("123456789", nil)

	v := url.Values{}
	v.Add("foo", "bar")
	v.Add("test", "this")

	res, err := testRequest(ts, "POST", "/apps/app/tables/table/rows", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(`"123456789"`), data)
}

func TestTableRowDelete(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	mp.On("TableRowDelete", "app", "table", "this", types.TableRowDeleteOptions{Index: "id"}).Return(nil)

	v := url.Values{}
	v.Add("app", "app")
	v.Add("table", "table")
	v.Add("index", "id")
	v.Add("key", "this")

	res, err := testRequest(ts, "DELETE", "/apps/app/tables/table/indexes/id/this", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(""), data)
}

func TestTableRowsDelete(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	mp.On("TableRowsDelete", "app", "table", []string{"app", "table"}, types.TableRowDeleteOptions{Index: "id"}).Return(nil)

	v := url.Values{}
	v.Add("key", "app")
	v.Add("key", "table")

	res, err := testRequest(ts, "POST", "/apps/app/tables/table/indexes/id/batch/remove", bytes.NewReader([]byte(v.Encode())))
	assert.NoError(t, err)
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []byte(""), data)
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
