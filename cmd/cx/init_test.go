package main_test

import (
	"testing"

	cx "github.com/convox/praxis/cmd/cx"
	//"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/manifest"
	mv1 "github.com/convox/rack/manifest"
	"github.com/stretchr/testify/assert"
)

func TestManifestConvert(t *testing.T) {
	m, err := mv1.LoadFile("testdata/docker-compose.yml")
	if !assert.NoError(t, err) {
		return
	}

	cm, err := cx.ManifestConvert(m)
	if !assert.NoError(t, err) {
		return
	}

	n := manifest.Manifest{
		Resources: manifest.Resources{
			manifest.Resource{
				Name: "database",
				Type: "postgres",
			},
		},
		Services: manifest.Services{
			manifest.Service{
				Name: "web",
				Build: manifest.ServiceBuild{
					Path: ".",
				},
				Command: manifest.ServiceCommand{
					Development: "bin/web",
					Production:  "bin/web",
					Test:        "",
				},
				Environment: []string{
					"BAZ",
					"FOO=bar",
					"QUX=",
				},
				Health: manifest.ServiceHealth{
					Path:     "/health",
					Interval: 5,
					Timeout:  60,
				},
				Image: "httpd",
				Port: manifest.ServicePort{
					Port:   80,
					Scheme: "http",
				},
				Resources: []string{"database"},
				Scale: manifest.ServiceScale{
					Memory: 50,
				},
				Volumes: []string{
					"/var/lib/postgresql/data",
					"/foo:/bar",
				},
			},
			manifest.Service{
				Name: "worker",
				Build: manifest.ServiceBuild{
					Path: ".",
				},
				Command: manifest.ServiceCommand{
					Development: "bin/work",
					Test:        "",
					Production:  "bin/work",
				},
				Environment: []string{},
				Health: manifest.ServiceHealth{
					Path:     "/",
					Interval: 5,
					Timeout:  4,
				},
				Port: manifest.ServicePort{
					Port:   0,
					Scheme: "",
				},
				Resources: []string{},
				Scale: manifest.ServiceScale{
					Memory: 256,
				},
			},
		},
		Timers: manifest.Timers{
			manifest.Timer{
				Name:     "myjob",
				Command:  "bin/myjob",
				Schedule: "30 18 ? * MON-FRI",
				Service:  "web",
			},
		},
	}

	assert.Equal(t, *cm, n)
}
