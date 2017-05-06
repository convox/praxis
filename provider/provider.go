package provider

import (
	"fmt"
	"os"

	"github.com/convox/praxis/provider/aws"
	"github.com/convox/praxis/provider/local"
	"github.com/convox/praxis/types"
)

// FromEnv returns a new Provider from env vars
func FromEnv() (types.Provider, error) {
	return FromType(os.Getenv("PROVIDER"))
}

func FromType(t string) (types.Provider, error) {
	switch t {
	case "aws":
		return aws.FromEnv()
	case "local":
		return local.FromEnv()
	default:
		return nil, fmt.Errorf("invalid provider type: %s", t)
	}
}
