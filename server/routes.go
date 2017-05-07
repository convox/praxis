package server

import (
	"net/http"
	"os"

	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
)

func Routes(server *api.Server) {
	server.Route("root", "GET", "/", func(w http.ResponseWriter, r *http.Request, c *api.Context) error {
		w.Write([]byte("ok"))
		return nil
	})

	auth := server.Subrouter("/")

	if pw := os.Getenv("PASSWORD"); pw != "" {
		auth.Use(authenticate(pw))
	}

	auth.Route("app.create", "POST", "/apps", controllers.AppCreate)
	auth.Route("app.delete", "DELETE", "/apps/{name}", controllers.AppDelete)
	auth.Route("app.get", "GET", "/apps/{name}", controllers.AppGet)
	auth.Route("app.list", "GET", "/apps", controllers.AppList)
	auth.Route("app.logs", "GET", "/apps/{app}/logs", controllers.AppLogs)
	auth.Route("app.registry", "GET", "/apps/{app}/registry", controllers.AppRegistry)

	auth.Route("build.create", "POST", "/apps/{app}/builds", controllers.BuildCreate)
	auth.Route("build.get", "GET", "/apps/{app}/builds/{id}", controllers.BuildGet)
	auth.Route("build.list", "GET", "/apps/{app}/builds", controllers.BuildList)
	auth.Route("build.logs", "GET", "/apps/{app}/builds/{id}/logs", controllers.BuildLogs)
	auth.Route("build.update", "PUT", "/apps/{app}/builds/{id}", controllers.BuildUpdate)

	auth.Route("files.delete", "DELETE", "/apps/{app}/processes/{process}/files", controllers.FilesDelete)
	auth.Route("files.upload", "POST", "/apps/{app}/processes/{process}/files", controllers.FilesUpload)

	auth.Route("key.decrypt", "POST", "/apps/{app}/keys/{key}/decrypt", controllers.KeyDecrypt)
	auth.Route("key.encrypt", "POST", "/apps/{app}/keys/{key}/encrypt", controllers.KeyEncrypt)

	auth.Route("object.exists", "HEAD", "/apps/{app}/objects/{key:.*}", controllers.ObjectExists)
	auth.Route("object.fetch", "GET", "/apps/{app}/objects/{key:.*}", controllers.ObjectFetch)
	auth.Route("object.store", "POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	auth.Stream("process.exec", "/apps/{app}/processes/{pid}/exec", controllers.ProcessExec)
	auth.Route("process.get", "GET", "/apps/{app}/processes/{pid}", controllers.ProcessGet)
	auth.Route("process.logs", "GET", "/apps/{app}/processes/{pid}/logs", controllers.ProcessLogs)
	auth.Route("process.list", "GET", "/apps/{app}/processes", controllers.ProcessList)
	auth.Stream("process.run", "/apps/{app}/processes/run", controllers.ProcessRun)
	auth.Route("process.start", "POST", "/apps/{app}/processes", controllers.ProcessStart)
	auth.Route("process.stop", "DELETE", "/apps/{app}/processes/{pid}", controllers.ProcessStop)

	auth.Route("proxy", "POST", "/apps/{app}/processes/{process}/proxy/{port}", controllers.Proxy)

	auth.Route("queue.fetch", "GET", "/apps/{app}/queues/{queue}", controllers.QueueFetch)
	auth.Route("queue.store", "POST", "/apps/{app}/queues/{queue}", controllers.QueueStore)

	auth.Route("registry.add", "POST", "/registries", controllers.RegistryAdd)
	auth.Route("registry.list", "GET", "/registries", controllers.RegistryList)
	auth.Route("registry.remove", "DELETE", "/registries/{hostname:.*}", controllers.RegistryRemove)

	auth.Route("release.create", "POST", "/apps/{app}/releases", controllers.ReleaseCreate)
	auth.Route("release.get", "GET", "/apps/{app}/releases/{id}", controllers.ReleaseGet)
	auth.Route("release.list", "GET", "/apps/{app}/releases", controllers.ReleaseList)
	auth.Route("release.logs", "GET", "/apps/{app}/releases/{id}/logs", controllers.ReleaseLogs)

	auth.Route("resource.list", "GET", "/apps/{app}/resources", controllers.ResourceList)

	auth.Route("service.list", "GET", "/apps/{app}/services", controllers.ServiceList)

	auth.Route("system.get", "GET", "/system", controllers.SystemGet)
	auth.Route("system.logs", "GET", "/system/logs", controllers.SystemLogs)
	auth.Route("system.options", "OPTIONS", "/system", controllers.SystemOptions)
	auth.Route("system.update", "POST", "/system", controllers.SystemUpdate)

	auth.Route("table.get", "GET", "/apps/{app}/tables/{table}", controllers.TableGet)
	auth.Route("table.list", "GET", "/apps/{app}/tables", controllers.TableList)
	auth.Route("table.truncate", "POST", "/apps/{app}/tables/{table}/truncate", controllers.TableTruncate)
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
