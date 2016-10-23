package local

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/convox/praxis/provider"
)

func (p *Provider) BlobExists(app, key string) (bool, error) {
	return p.exists(fmt.Sprintf("blob/%s/%s", app, key))
}

func (p *Provider) BlobFetch(app, key string) (io.ReadCloser, error) {
	return p.load(fmt.Sprintf("blob/%s/%s", app, key))
}

func (p *Provider) BlobStore(app, key string, r io.Reader, opts provider.BlobStoreOptions) (string, error) {
	if key == "" {
		tmp, err := generateTempKey()
		if err != nil {
			return "", err
		}

		key = tmp
	}

	if err := p.save(fmt.Sprintf("blob/%s/%s", app, key), r); err != nil {
		return "", err
	}

	return fmt.Sprintf("blob:///%s", key), nil
}

func generateTempKey() (string, error) {
	data := make([]byte, 1024)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)

	return fmt.Sprintf("tmp/%s", hex.EncodeToString(hash[:])[0:30]), nil
}
