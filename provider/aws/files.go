package aws

import (
	"io"
)

func (p *Provider) FilesDelete(app, pid string, files []string) error {
	return nil
}

func (p *Provider) FilesUpload(app, pid string, r io.Reader) error {
	return nil
}
