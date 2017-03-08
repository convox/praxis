package provider

import (
	"io"
	"os"

	"github.com/convox/praxis/provider/local"
	"github.com/convox/praxis/types"
)

type Provider interface {
	// Initialize(opts structs.ProviderOptions) error

	// AppCancel(name string) error
	AppCreate(name string) (*types.App, error)
	AppDelete(name string) error
	AppGet(name string) (*types.App, error)
	AppList() (types.Apps, error)

	BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error)
	// BuildDelete(app, id string) (*structs.Build, error)
	// BuildExport(app, id string, w io.Writer) error
	BuildGet(app, id string) (*types.Build, error)
	// BuildImport(app string, r io.Reader) (*structs.Build, error)
	BuildLogs(app, id string) (io.ReadCloser, error)
	// BuildList(app string, limit int64) (structs.Builds, error)
	// BuildRelease(*structs.Build) (*structs.Release, error)
	BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error)

	// CapacityGet() (*structs.Capacity, error)

	// CertificateCreate(pub, key, chain string) (*structs.Certificate, error)
	// CertificateDelete(id string) error
	// CertificateGenerate(domains []string) (*structs.Certificate, error)
	// CertificateList() (structs.Certificates, error)

	// EventSend(*structs.Event, error) error

	// EnvironmentGet(app string) (structs.Environment, error)

	FilesDelete(app, pid string, files []string) error
	FilesUpload(app, pid string, r io.Reader) error

	// FormationList(app string) (structs.Formation, error)
	// FormationGet(app, process string) (*structs.ProcessFormation, error)
	// FormationSave(app string, pf *structs.ProcessFormation) error

	// InstanceList() (structs.Instances, error)
	// InstanceTerminate(id string) error

	KeyDecrypt(app, key string, data []byte) ([]byte, error)
	KeyEncrypt(app, key string, data []byte) ([]byte, error)

	// LogStream(app string, w io.Writer, opts structs.LogStreamOptions) error

	// ObjectDelete(key string) error
	// ObjectExists(key string) bool
	ObjectFetch(app, key string) (io.ReadCloser, error)
	// ObjectList(prefix string) ([]string, error)
	ObjectStore(app, key string, r io.Reader, opts types.ObjectStoreOptions) (*types.Object, error)

	// ProcessExec(app, pid, command string, stream io.ReadWriter, opts structs.ProcessExecOptions) error
	ProcessGet(app, pid string) (*types.Process, error)
	ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error)
	ProcessLogs(app, pid string) (io.ReadCloser, error)
	ProcessRun(app string, opts types.ProcessRunOptions) (int, error)
	ProcessStart(app string, opts types.ProcessStartOptions) (string, error)
	ProcessStop(app, pid string) error

	Proxy(app, pid string, port int, in io.Reader) (io.ReadCloser, error)

	// RegistryAdd(server, username, password string) (*structs.Registry, error)
	// RegistryDelete(server string) error
	// RegistryList() (structs.Registries, error)

	QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error)
	QueueStore(app, queue string, attrs map[string]string) error

	ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error)
	ReleaseGet(app, id string) (*types.Release, error)
	ReleaseList(app string) (types.Releases, error)
	ReleasePromote(app, id string) error

	SystemGet() (*types.System, error)
	// SystemLogs(w io.Writer, opts structs.LogStreamOptions) error
	// SystemProcesses(opts structs.SystemProcessesOptions) (structs.Processes, error)
	// SystemSave(system structs.System) error

	// TableCreate(app, name string, opts types.TableCreateOptions) error
	TableFetch(app, table, key string, opts types.TableFetchOptions) (map[string]string, error)
	TableFetchBatch(app, table string, key []string, opts types.TableFetchOptions) ([]map[string]string, error)
	TableGet(app, table string) (*types.Table, error)
	TableList(app string) (types.Tables, error)
	TableStore(app, table string, attrs map[string]string) (string, error)
	TableTruncate(app, table string) error
}

// FromEnv returns a new Provider from env vars
func FromEnv() Provider {
	switch os.Getenv("PROVIDER") {
	// case "aws":
	//   return aws.FromEnv()
	case "test":
		return &MockProvider{}
	default:
		return local.FromEnv()
	}
}
