package local

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	"github.com/convox/logger"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=provider.local")

type Provider struct {
	Frontend string
	Name     string
	Root     string
	Test     bool
	Version  string

	db *bolt.DB
}

// FromEnv returns a new local.Provider from env vars
func FromEnv() (*Provider, error) {
	p := &Provider{
		Name:     coalesce(os.Getenv("NAME"), "convox"),
		Frontend: coalesce(os.Getenv("PROVIDER_FRONTEND"), "10.42.84.0"),
		Root:     coalesce(os.Getenv("PROVIDER_ROOT"), "/var/convox"),
		Version:  "latest",
	}

	vf := "/var/convox/version"

	if _, err := os.Stat(vf); !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(vf)
		if err != nil {
			return nil, err
		}

		p.Version = strings.TrimSpace(string(data))
	}

	if err := p.Init(); err != nil {
		return nil, err
	}

	return p, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func (p *Provider) Init() error {
	// if os.Getenv("PROVIDER_LOCAL_SKIP_FRONTEND_CHECK") != "true" {
	//   if err := p.checkFrontend(); err != nil {
	//     return err
	//   }
	// }

	if err := os.MkdirAll(p.Root, 0700); err != nil {
		return err
	}

	db, err := bolt.Open(filepath.Join(p.Root, "rack.db"), 0600, nil)
	if err != nil {
		return err
	}

	p.db = db

	if _, err := p.createRootBucket("rack"); err != nil {
		return err
	}

	go p.workers()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		p.shutdown()
	}()

	return nil
}

// shutdown cleans up any running resources and exit
func (p *Provider) shutdown() {
	cs, err := containersByLabels(map[string]string{
		"convox.rack": p.Name,
	})
	if err != nil {
		fmt.Printf("shutdown error %s", err)
		return
	}

	for _, id := range cs {
		if err := exec.Command("docker", "kill", id).Run(); err != nil {
			fmt.Printf("shutdown error %s", err)
		}
	}

	os.Exit(1)
}

func (p *Provider) createRootBucket(name string) (*bolt.Bucket, error) {
	tx, err := p.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bucket, err := tx.CreateBucketIfNotExists([]byte("rack"))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return bucket, err
}

func (p *Provider) checkFrontend() error {
	if p.Frontend == "none" {
		return nil
	}

	c := http.DefaultClient

	c.Timeout = 2 * time.Second

	if _, err := c.Get(fmt.Sprintf("http://%s:9477/endpoints", p.Frontend)); err != nil {
		return fmt.Errorf("unable to register with frontend")
	}

	return nil
}
