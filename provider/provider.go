package provider

import (
	"io"

	"github.com/convox/praxis/provider/models"
)

type Provider interface {
	AppCreate(name string, opts models.AppCreateOptions) (*models.App, error)

	BlobStore(app, key string, r io.Reader, opts models.BlobStoreOptions) (string, error)

	BuildCreate(app, url string, opts models.BuildCreateOptions) (*models.Build, error)

	ProcessRun(app, service string, opts models.ProcessRunOptions) (*models.Process, error)
}
