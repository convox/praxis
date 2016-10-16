package local

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/convox/praxis/provider/models"
)

func (p *Provider) BlobStore(key string, r io.Reader, opts models.BlobStoreOptions) (string, error) {
	fmt.Printf("key = %+v\n", key)
	if key == "" {
		tmp, err := generateTempKey()
		if err != nil {
			return "", err
		}

		key = tmp
	}

	if err := p.save(fmt.Sprintf("blob/%s", key), r); err != nil {
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
