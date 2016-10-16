package api

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"time"

	"github.com/convox/logger"
	"github.com/convox/nlogger"
	"github.com/convox/praxis/api/routes"
	"github.com/urfave/negroni"
)

func Listen(addr string) error {
	log := logger.New("ns=api").At("start")

	r := routes.New()

	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(nlogger.New("ns=api", nil))
	n.UseHandler(r)

	log.Logf("addr=%s", addr)

	pub, key, err := generateCertificate("localhost")
	if err != nil {
		log.Error(err)
		return err
	}

	cert, err := tls.X509KeyPair(pub, key)
	if err != nil {
		log.Error(err)
		return err
	}

	server := http.Server{
		Addr:    addr,
		Handler: n,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func generateCertificate(host string) ([]byte, []byte, error) {
	rkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"convox"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	data, err := x509.CreateCertificate(rand.Reader, &template, &template, &rkey.PublicKey, rkey)
	if err != nil {
		return nil, nil, err
	}

	pub := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data})
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rkey)})

	return pub, key, nil
}
