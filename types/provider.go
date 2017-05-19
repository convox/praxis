package types

import (
	"context"
	"io"
)

type Provider interface {
	AppCreate(name string) (*App, error)
	AppDelete(name string) error
	AppGet(name string) (*App, error)
	AppList() (Apps, error)
	AppLogs(app string, opts LogsOptions) (io.ReadCloser, error)
	AppRegistry(app string) (*Registry, error)

	BuildCreate(app, url string, opts BuildCreateOptions) (*Build, error)
	// BuildExport(app, id string, w io.Writer) error
	BuildGet(app, id string) (*Build, error)
	// BuildImport(app string, r io.Reader) (*structs.Build, error)
	BuildLogs(app, id string) (io.ReadCloser, error)
	BuildList(app string) (Builds, error)
	BuildUpdate(app, id string, opts BuildUpdateOptions) (*Build, error)

	CacheFetch(app, cache, key string) (map[string]string, error)
	CacheStore(app, cache, key string, attrs map[string]string, opts CacheStoreOptions) error

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
	ObjectStore(app, key string, r io.Reader, opts ObjectStoreOptions) (*Object, error)

	ProcessExec(app, pid, command string, opts ProcessExecOptions) (int, error)
	ProcessGet(app, pid string) (*Process, error)
	ProcessList(app string, opts ProcessListOptions) (Processes, error)
	ProcessLogs(app, pid string, opts LogsOptions) (io.ReadCloser, error)
	ProcessRun(app string, opts ProcessRunOptions) (int, error)
	ProcessStart(app string, opts ProcessRunOptions) (string, error)
	ProcessStop(app, pid string) error

	Proxy(app, pid string, port int, in io.Reader) (io.ReadCloser, error)

	QueueFetch(app, queue string, opts QueueFetchOptions) (map[string]string, error)
	QueueStore(app, queue string, attrs map[string]string) error

	RegistryAdd(server, username, password string) (*Registry, error)
	RegistryList() (Registries, error)
	RegistryRemove(server string) error

	ReleaseCreate(app string, opts ReleaseCreateOptions) (*Release, error)
	ReleaseGet(app, id string) (*Release, error)
	ReleaseList(app string, opts ReleaseListOptions) (Releases, error)
	ReleaseLogs(app, id string, opts LogsOptions) (io.ReadCloser, error)
	ReleasePromote(app, id string) error

	ResourceList(app string) (Resources, error)

	ServiceList(app string) (Services, error)

	SystemGet() (*System, error)
	SystemInstall(name string, opts SystemInstallOptions) (string, error)
	SystemLogs(opts LogsOptions) (io.ReadCloser, error)
	SystemOptions() (map[string]string, error)
	SystemUninstall(name string, opts SystemInstallOptions) error
	SystemUpdate(opts SystemUpdateOptions) error

	TableGet(app, table string) (*Table, error)
	TableList(app string) (Tables, error)
	TableQuery(app, table, query string) (TableRows, error)
	TableTruncate(app, table string) error

	WithContext(ctx context.Context) Provider

	Workers()
}
