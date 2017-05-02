package aws

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/secretbox"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

const (
	keyLength   = 32
	nonceLength = 24
)

type keyEnvelope struct {
	CipherText   []byte `json:"c"`
	EncryptedKey []byte `json:"k"`
	Nonce        []byte `json:"n"`
}

func (p *Provider) KeyDecrypt(app, key string, data []byte) ([]byte, error) {
	k, err := p.appResource(app, fmt.Sprintf("Key%s", upperName(key)))
	if err != nil {
		return nil, err
	}

	dd, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	var e keyEnvelope

	if err := json.Unmarshal(dd, &e); err != nil {
		return nil, err
	}

	res, err := p.KMS().Decrypt(&kms.DecryptInput{
		CiphertextBlob: e.EncryptedKey,
	})
	if err != nil {
		return nil, err
	}

	parts := strings.Split(*res.KeyId, "/")

	if parts[len(parts)-1] != k {
		return nil, fmt.Errorf("incorrect key")
	}

	var dk [keyLength]byte
	copy(dk[:], res.Plaintext[0:keyLength])

	var n [nonceLength]byte
	copy(n[:], e.Nonce[0:nonceLength])

	var dec []byte
	dec, ok := secretbox.Open(dec, e.CipherText, &n, &dk)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}

	return dec, nil
}

func (p *Provider) KeyEncrypt(app, key string, data []byte) ([]byte, error) {
	k, err := p.appResource(app, fmt.Sprintf("Key%s", upperName(key)))
	if err != nil {
		return nil, err
	}

	res, err := p.KMS().GenerateDataKey(&kms.GenerateDataKeyInput{
		KeyId:         aws.String(k),
		NumberOfBytes: aws.Int64(keyLength),
	})
	if err != nil {
		return nil, err
	}

	var dk [keyLength]byte
	copy(dk[:], res.Plaintext[0:keyLength])

	rnd, err := p.generateRandom()
	if err != nil {
		return nil, err
	}

	var n [nonceLength]byte
	copy(n[:], rnd[0:nonceLength])

	var enc []byte
	enc = secretbox.Seal(enc, data, &n, &dk)

	e := keyEnvelope{
		CipherText:   enc,
		EncryptedKey: res.CiphertextBlob,
		Nonce:        n[:],
	}

	ed, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	ee := base64.StdEncoding.EncodeToString(ed)

	return []byte(ee), nil
}

func (p *Provider) generateRandom() ([]byte, error) {
	res, err := p.KMS().GenerateRandom(&kms.GenerateRandomInput{NumberOfBytes: aws.Int64(nonceLength)})

	if err != nil {
		return nil, err
	}

	return res.Plaintext, nil
}
