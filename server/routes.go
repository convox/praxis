package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
)

func Routes(server *api.Server) {
	server.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	auth := server.Subrouter("/")

	if pw := os.Getenv("PASSWORD"); pw != "" {
		auth.Use(authenticate(pw))
	}

	auth.Route("POST", "/apps", controllers.AppCreate)
	auth.Route("DELETE", "/apps/{name}", controllers.AppDelete)
	auth.Route("GET", "/apps/{name}", controllers.AppGet)
	auth.Route("GET", "/apps", controllers.AppList)
	auth.Route("GET", "/apps/{app}/logs", controllers.AppLogs)
	auth.Route("GET", "/apps/{app}/registry", controllers.AppRegistry)

	auth.Route("POST", "/apps/{app}/builds", controllers.BuildCreate)
	auth.Route("GET", "/apps/{app}/builds/{id}", controllers.BuildGet)
	auth.Route("GET", "/apps/{app}/builds", controllers.BuildList)
	auth.Route("GET", "/apps/{app}/builds/{id}/logs", controllers.BuildLogs)
	auth.Route("PUT", "/apps/{app}/builds/{id}", controllers.BuildUpdate)

	auth.Route("GET", "/apps/{app}/caches/{cache}/{key}", controllers.CacheFetch)
	auth.Route("POST", "/apps/{app}/caches/{cache}/{key}", controllers.CacheStore)

	auth.Route("DELETE", "/apps/{app}/processes/{process}/files", controllers.FilesDelete)
	auth.Route("POST", "/apps/{app}/processes/{process}/files", controllers.FilesUpload)

	auth.Route("POST", "/apps/{app}/keys/{key}/decrypt", controllers.KeyDecrypt)
	auth.Route("POST", "/apps/{app}/keys/{key}/encrypt", controllers.KeyEncrypt)

	auth.Route("GET", "/apps/{app}/metrics/{name}", controllers.MetricGet)
	auth.Route("GET", "/apps/{app}/metrics", controllers.MetricList)

	auth.Route("HEAD", "/apps/{app}/objects/{key:.*}", controllers.ObjectExists)
	auth.Route("GET", "/apps/{app}/objects/{key:.*}", controllers.ObjectFetch)
	auth.Route("POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	auth.Stream("process.exec", "/apps/{app}/processes/{pid}/exec", controllers.ProcessExec)
	auth.Stream("process.run", "/apps/{app}/processes/run", controllers.ProcessRun)
	auth.Route("GET", "/apps/{app}/processes/{pid}", controllers.ProcessGet)
	auth.Route("GET", "/apps/{app}/processes/{pid}/logs", controllers.ProcessLogs)
	auth.Route("GET", "/apps/{app}/processes", controllers.ProcessList)
	auth.Stream("process.proxy", "/apps/{app}/processes/{pid}/proxy/{port}", controllers.ProcessProxy)
	auth.Route("POST", "/apps/{app}/processes", controllers.ProcessStart)
	auth.Route("DELETE", "/apps/{app}/processes/{pid}", controllers.ProcessStop)

	auth.Route("GET", "/apps/{app}/queues/{queue}", controllers.QueueFetch)
	auth.Route("POST", "/apps/{app}/queues/{queue}", controllers.QueueStore)

	auth.Route("POST", "/registries", controllers.RegistryAdd)
	auth.Route("GET", "/registries", controllers.RegistryList)
	auth.Route("DELETE", "/registries/{hostname:.*}", controllers.RegistryRemove)

	auth.Route("POST", "/apps/{app}/releases", controllers.ReleaseCreate)
	auth.Route("GET", "/apps/{app}/releases/{id}", controllers.ReleaseGet)
	auth.Route("GET", "/apps/{app}/releases", controllers.ReleaseList)
	auth.Route("GET", "/apps/{app}/releases/{id}/logs", controllers.ReleaseLogs)
	auth.Route("POST", "/apps/{app}/releases/{id}", controllers.ReleasePromote)

	auth.Stream("resource.proxy", "/apps/{app}/resources/{name}/proxy", controllers.ResourceProxy)
	auth.Route("GET", "/apps/{app}/resources/{name}", controllers.ResourceGet)
	auth.Route("GET", "/apps/{app}/resources", controllers.ResourceList)

	auth.Route("GET", "/apps/{app}/services/{name}", controllers.ServiceGet)
	auth.Route("GET", "/apps/{app}/services", controllers.ServiceList)

	// auth.Stream("system.proxy", "/system/proxy/{host}/{port}", controllers.SystemProxy)
	auth.Route("GET", "/system", controllers.SystemGet)
	auth.Route("GET", "/system/logs", controllers.SystemLogs)
	auth.Route("OPTIONS", "/system", controllers.SystemOptions)
	auth.Route("POST", "/system", controllers.SystemUpdate)

	// pprof
	auth.Router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	auth.Router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	auth.Router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	auth.Router.HandleFunc("/debug/pprof/{topic:.*}", pprof.Index)
}

func authenticate(password string) api.Middleware {
	return func(fn api.HandlerFunc) api.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, c *api.Context) error {
			key, _, ok := r.BasicAuth()

			if !ok || key != password {
				return api.Errorf(401, "invalid auth")
			}

			return fn(w, r, c)
		}
	}
}
