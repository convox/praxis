package controllers

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"

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
