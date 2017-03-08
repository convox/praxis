package controllers_test

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestTableFetchBatch(t *testing.T) {
	ts, mp := mockServer()
	defer ts.Close()

	items := []map[string]string{
		map[string]string{"id": "one", "foo": "bar1"},
		map[string]string{"id": "two", "foo": "bar2"},
	}

	mp.On("TableFetchBatch", "app", "table", []string{"one", "two"}, types.TableFetchOptions{Index: "id"}).Return(items, nil)

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
