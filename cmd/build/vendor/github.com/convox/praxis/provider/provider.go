package provider

import (
	"io"

	"github.com/convox/praxis/provider/models"
)

type Provider interface {
	AppCreate(name string, opts models.AppCreateOptions) (*models.App, error)

	BlobStore(app, key string, r io.Reader, opts models.BlobStoreOptions) (string, error)

	BuildCreate(app, url string, opts models.BuildCreateOptions) (*models.Build, error)
	BuildLoad(app, id string) (*models.Build, error)
	BuildSave(build *models.Build) error

	ProcessStart(app, service string, opts models.ProcessRunOptions) (*models.Process, error)
	ProcessWait(app, pid string) (int, error)

	TableLoad(app, table, id string) (map[string]string, error)
	TableSave(app, table, id string, attrs map[string]string) error
}
