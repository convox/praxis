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

	cm, report, err := cx.ManifestConvert(m)
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

	r := cx.Report{
		Messages: []string{
			"<fail>FAIL</fail>: <service>web</service> build args not migrated to convox.yml, use ARG in your Dockerfile instead\n",
			"<fail>FAIL</fail>: <service>web</service> \"dockerfile\" key is not supported in convox.yml, file must be named \"Dockerfile\"\n",
			"<fail>FAIL</fail>: <service>web</service> \"entrypoint\" key not supported in convox.yml, use ENTRYPOINT in Dockerfile instead\n",
			"INFO: <service>web</service> - running as an agent is not supported\n",
			"INFO: <service>web</service> - setting draning timeout is not supported\n",
			"INFO: <service>web</service> - setting secure environment is not necessary\n",
			"INFO: <service>web</service> - setting health check port is not necessary\n",
			"INFO: <service>web</service> - setting health check thresholds is not supported\n",
			"INFO: <service>web</service> - setting idle timeout is not supported\n",
			"INFO: <service>web</service> - configuring balancer via convox.port labels is not supported\n",
			"<fail>FAIL</fail>: <service>web</service> - port shifting is not supported, use internal hostnames instead\n",
			"INFO: <service>web</service> - environment variables not generated for linked service <service>worker</service>, use internal URL https://worker.<app name>.convox instead\n",
			"INFO: <service>web</service> - multiple ports found, only 1 HTTP port per service is supported\n",
			"INFO: <service>web</service> - only HTTP ports supported\n",
			"INFO: <service>web</service> - UDP ports are not supported\n",
			"INFO: <service>web</service> - only HTTP ports supported\n",
			"INFO: <service>web</service> - privileged mode not supported\n",
			"INFO: <service>database</service> has been migrated to a resource\n",
			"INFO: custom networks not supported, use service hostnames instead\n",
		},
		Success: false,
	}

	assert.Equal(t, *cm, n)
	assert.Equal(t, report, r)
}
