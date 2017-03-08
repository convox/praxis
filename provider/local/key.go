package local

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

const (
	aesKey   = "AES256Key-32Characters1234567890"
	nonceHex = "37b8e8a308c354048d245f6d"
)

func (p *Provider) KeyDecrypt(app, key string, data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil, err
	}

	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, nonce, data, nil)
}

func (p *Provider) KeyEncrypt(app, key string, data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil, err
	}

	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, nonce, data, nil), nil
}
