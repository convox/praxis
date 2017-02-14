package controllers

import "github.com/convox/praxis/provider"

var (
	Provider provider.Provider
)

func init() {
	Provider = provider.FromEnv()
}
