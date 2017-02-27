package local

import "io"

func (p *Provider) ProxyStart(app, pid string, port int) (io.ReadWriter, error) {
	_, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
