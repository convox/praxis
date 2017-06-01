package local

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/pkg/errors"
)

const (
	aesKey   = "AES256Key-32Characters1234567890"
	nonceHex = "37b8e8a308c354048d245f6d"
)

func (p *Provider) KeyDecrypt(app, key string, data []byte) ([]byte, error) {
	log := p.logger("KeyDecrypt").Append("app=%q key=%q", app, key)

	if _, err := p.AppGet(app); err != nil {
		return nil, log.Error(err)
	}

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	dec, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return dec, log.Success()
}

func (p *Provider) KeyEncrypt(app, key string, data []byte) ([]byte, error) {
	log := p.logger("KeyEncrypt").Append("app=%q key=%q", app, key)

	if _, err := p.AppGet(app); err != nil {
		return nil, log.Error(err)
	}

	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	enc, err := aesgcm.Seal(nil, nonce, data, nil), nil
	if err != nil {
		return nil, errors.WithStack(log.Error(err))
	}

	return enc, log.Success()
}
