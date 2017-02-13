package manifest_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/convox/praxis/manifest"
	"github.com/stretchr/testify/assert"
)

func TestManifestLoad(t *testing.T) {
	data, err := testdata("full")
	if !assert.NoError(t, err) {
		return
	}

	m, err := manifest.Load(data)
	if !assert.NoError(t, err) {
		return
	}

	n := &manifest.Manifest{
		Balancers: manifest.Balancers{
			manifest.Balancer{
				Name: "api",
				Endpoints: manifest.BalancerEndpoints{
					manifest.BalancerEndpoint{Port: "80", Protocol: "http", Redirect: "https://:443"},
					manifest.BalancerEndpoint{Port: "443", Protocol: "https", Target: "http://api:3000"},
				},
			},
			manifest.Balancer{
				Name: "proxy",
				Endpoints: manifest.BalancerEndpoints{
					manifest.BalancerEndpoint{Port: "80", Target: "proxy:3001"},
					manifest.BalancerEndpoint{Port: "443", Target: "proxy:3002"},
					manifest.BalancerEndpoint{Port: "1080", Target: "proxy:3003"},
					manifest.BalancerEndpoint{Port: "2133", Target: "proxy:3000"},
				},
			},
		},
		Queues: manifest.Queues{
			manifest.Queue{
				Name:    "traffic",
				Timeout: "5m",
			},
		},
		Services: manifest.Services{
			manifest.Service{
				Name: "api",
				Build: manifest.ServiceBuild{
					Path: "api",
				},
				Environment: []string{
					"DEVELOPMENT=false",
					"SECRET",
				},
			},
			manifest.Service{
				Name:  "proxy",
				Image: "ubuntu:16.04",
				Environment: []string{
					"SECRET",
				},
			},
		},
		Tables: manifest.Tables{
			manifest.Table{
				Name: "proxies",
				Indexes: []string{
					"password",
					"password-previous",
					"heroku-id",
				},
			},
			manifest.Table{
				Name: "servers",
				Indexes: []string{
					"ip",
					"type-available",
				},
			},
			manifest.Table{
				Name: "stats",
				Indexes: []string{
					"proxy-month",
				},
			},
			manifest.Table{
				Name: "traffic",
				Indexes: []string{
					"proxy:started",
				},
			},
		},
	}

	assert.Equal(t, n, m)
}

func TestManifestLoadInvalid(t *testing.T) {
	m, err := testdataManifest("invalid.1")
	assert.Nil(t, m)
	assert.Error(t, err, "yaml: line 2: did not find expected comment or line break")

	m, err = testdataManifest("invalid.2")
	assert.Nil(t, m)
	assert.Error(t, err, "yaml: line 3: mapping values are not allowed in this context")
}

func testdata(name string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("testdata/%s.yml", name))
}

func testdataManifest(name string) (*manifest.Manifest, error) {
	data, err := testdata(name)
	if err != nil {
		return nil, err
	}

	return manifest.Load(data)
}
