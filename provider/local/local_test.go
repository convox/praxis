package local_test

import (
	"io/ioutil"

	"github.com/convox/praxis/provider/local"
)

func Provider() (*local.Provider, error) {
	tmp, err := ioutil.TempDir("", "praxis")
	if err != nil {
		return nil, err
	}

	return &local.Provider{Root: tmp}, nil
}
