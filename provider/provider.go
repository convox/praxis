package provider

import (
	"fmt"
	"os"

	"github.com/convox/praxis/provider/local"
	"github.com/convox/praxis/provider/models"
)

type Provider interface {
	TableList() (models.Tables, error)
}

func FromEnv() Provider {
	switch os.Getenv("PROVIDER") {
	case "local":
		return local.FromEnv()
	case "test":
		return &MockProvider{}
	default:
		panic(fmt.Errorf("unknown PROVIDER"))
	}
}
