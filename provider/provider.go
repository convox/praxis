package provider

import (
	"fmt"
	"os"

	"github.com/convox/praxis/models"
	"github.com/convox/praxis/provider/local"
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
