package local

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/convox/praxis/logger"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=p.local")

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Provider struct {
	Frontend string
	Name     string
	Root     string
	Test     bool
	Version  string

	ctx context.Context
	db  *bolt.DB
}

func FromEnv() (*Provider, error) {
	p := &Provider{
		Name:     coalesce(os.Getenv("NAME"), "convox"),
		Frontend: coalesce(os.Getenv("PROVIDER_FRONTEND"), "10.42.84.0"),
		Root:     coalesce(os.Getenv("PROVIDER_ROOT"), "/var/convox"),
		Version:  "latest",
	}

	if v := os.Getenv("VERSION"); v != "" {
		p.Version = v
	}

	if err := p.Init(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Provider) Context() context.Context {
	if p.ctx != nil {
		return p.ctx
	}

	return context.Background()
}

func (p *Provider) Init() error {
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

	return nil
}

func (p *Provider) logger(at string) *logger.Logger {
	if p.Test {
		return logger.NewWriter("", ioutil.Discard)
	}

	log := logger.New("ns=local")

	if id := p.Context().Value("request.id"); id != nil {
		log = log.Prepend("id=%s", id)
	}

	return log.At(at).Start()
}

// shutdown cleans up any running resources and exit
func (p *Provider) shutdown() error {
	cs, err := containersByLabels(map[string]string{
		"convox.rack": p.Name,
	})
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, id := range cs {
		wg.Add(1)
		go p.containerStopAsync(id, &wg)
	}

	wg.Wait()

	os.Exit(0)

	return nil
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
