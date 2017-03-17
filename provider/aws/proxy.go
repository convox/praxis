package aws

import (
	"io"
)

func (p *Provider) Proxy(app, pid string, port int, in io.Reader) (io.ReadCloser, error) {
	return nil, nil
}
