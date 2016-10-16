package provider

import (
	"io"

	"github.com/convox/praxis/provider/models"
)

type Provider interface {
	BlobStore(key string, r io.Reader, opts models.BlobStoreOptions) (string, error)

	BuildCreate(url string, opts models.BuildCreateOptions) (*models.Build, error)

	ProcessRun(service string, opts models.ProcessRunOptions) (*models.Process, error)
}
