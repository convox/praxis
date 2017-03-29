package controllers

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"

	"github.com/convox/praxis/provider"
)

const (
	sortableTime = "20060102.150405.000000000"
)

var (
	Provider provider.Provider
)

func Init() error {
	p, err := provider.FromEnv()
	if err != nil {
		return err
	}

	Provider = p

	return nil
}

func randomString() (string, error) {
	rb := make([]byte, 128)

	if _, err := rand.Read(rb); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha1.Sum(rb)), nil
}

func stream(w io.Writer, r io.Reader) error {
	buf := make([]byte, 1024)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, err := w.Write(buf[0:n]); err != nil {
				return err
			}
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
