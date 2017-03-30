package aws

import "fmt"

func (p *Provider) KeyDecrypt(app, key string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) KeyEncrypt(app, key string, data []byte) ([]byte, error) {
	return nil, fmt.Errorf("unimplemented")
}
