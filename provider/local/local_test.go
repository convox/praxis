package local_test

import (
	"io/ioutil"
	"os"

	"github.com/convox/praxis/provider/local"
)

func testProvider() (*local.Provider, error) {
	tmp, err := ioutil.TempDir("", "praxis")
	if err != nil {
		return nil, err
	}

	p := &local.Provider{
		Frontend: "none",
		Root:     tmp,
		Test:     true,
	}

	if err := p.Init(); err != nil {
		return nil, err
	}

	return p, nil
}

func testProviderCleanup(p *local.Provider) {
	if p.Root != "" {
		os.RemoveAll(p.Root)
	}
}
