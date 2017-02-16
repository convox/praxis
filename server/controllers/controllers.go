package controllers

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"

	"github.com/convox/praxis/provider"
)

var (
	Provider provider.Provider
)

func init() {
	Provider = provider.FromEnv()
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
		if err != nil && err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if _, err := w.Write(buf[0:n]); err != nil {
			return err
		}

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
