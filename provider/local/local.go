package local

import (
	"math/rand"
	"os"
	"time"

	"github.com/convox/logger"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	customTopic       = os.Getenv("CUSTOM_TOPIC")
	notificationTopic = os.Getenv("NOTIFICATION_TOPIC")
	sortableTime      = "20060102.150405.000000000"
)

// Logger is a package-wide logger
var Logger = logger.New("ns=provider.aws")

type Provider struct {
	Root string
}

// NewProviderFromEnv returns a new AWS provider from env vars
func FromEnv() *Provider {
	home, err := homedir.Expand("~/.convox/rack")
	if err != nil {
		panic(err)
	}

	return &Provider{Root: home}
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
