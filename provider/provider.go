package provider

import (
	"fmt"
	"os"

	"github.com/convox/praxis/provider/aws"
	"github.com/convox/praxis/provider/local"
	"github.com/convox/praxis/types"
)

type Provider interface {
	// AppCancel(name string) error
	AppCreate(name string) (*types.App, error)
	AppDelete(name string) error
	AppGet(name string) (*types.App, error)
	AppList() (types.Apps, error)
	AppLogs(app string, opts types.LogsOptions) (io.ReadCloser, error)
	AppRegistry(app string) (*types.Registry, error)

	BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error)
	// BuildExport(app, id string, w io.Writer) error
	BuildGet(app, id string) (*types.Build, error)
	// BuildImport(app string, r io.Reader) (*structs.Build, error)
	BuildLogs(app, id string) (io.ReadCloser, error)
	BuildList(app string) (types.Builds, error)
	BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error)

	FilesDelete(app, pid string, files []string) error
	FilesUpload(app, pid string, r io.Reader) error

	// InstanceList() (structs.Instances, error)
	// InstanceTerminate(id string) error

	KeyDecrypt(app, key string, data []byte) ([]byte, error)
	KeyEncrypt(app, key string, data []byte) ([]byte, error)

	// ObjectDelete(key string) error
	ObjectExists(app, key string) (bool, error)
	ObjectFetch(app, key string) (io.ReadCloser, error)
	// ObjectList(prefix string) ([]string, error)
	ObjectStore(app, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error)

	ProcessExec(app, pid, command string, opts types.ProcessExecOptions) (int, error)
	ProcessGet(app, pid string) (*types.Process, error)
	ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error)
	ProcessLogs(app, pid string, opts types.LogsOptions) (io.ReadCloser, error)
	ProcessRun(app string, opts types.ProcessRunOptions) (int, error)
	ProcessStart(app string, opts types.ProcessRunOptions) (string, error)
	ProcessStop(app, pid string) error

	Proxy(app, pid string, port int, in io.Reader) (io.ReadCloser, error)

	QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error)
	QueueStore(app, queue string, attrs map[string]string) error

	RegistryAdd(server, username, password string) (*types.Registry, error)
	RegistryList() (types.Registries, error)
	RegistryRemove(server string) error

	ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error)
	ReleaseGet(app, id string) (*types.Release, error)
	ReleaseList(app string, opts types.ReleaseListOptions) (types.Releases, error)
	ReleaseLogs(app, id string, opts types.LogsOptions) (io.ReadCloser, error)

	ResourceList(app string) (types.Resources, error)

	ServiceList(app string) (types.Services, error)

	SystemGet() (*types.System, error)
	SystemInstall(name string, opts types.SystemInstallOptions) (string, error)
	SystemLogs(opts types.LogsOptions) (io.ReadCloser, error)
	SystemUninstall(name string, opts types.SystemInstallOptions) error
	SystemUpdate(opts types.SystemUpdateOptions) error

	TableGet(app, table string) (*types.Table, error)
	TableList(app string) (types.Tables, error)
	TableQuery(app, table, query string) (types.TableRows, error)
	TableTruncate(app, table string) error
}

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
