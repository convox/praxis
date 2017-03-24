package local

import (
	"io"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/convox/logger"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=provider.aws")

type Provider struct {
	Test bool
	Root string
}

// FromEnv returns a new local.Provider from env vars
func FromEnv() (*Provider, error) {
	return &Provider{Root: coalesce(os.Getenv("PROVIDER_ROOT"), "/var/convox")}, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (p *Provider) Logs(pid string) (io.ReadCloser, error) {
	r, w := io.Pipe()

	cmd := exec.Command("docker", "logs", "--follow", pid)

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		w.Close()
	}()

	return r, nil
}
