package server

import (
	"github.com/convox/api"
	"github.com/convox/praxis/server/controllers"
)

func Routes(server *api.Server) {
	server.Route("app.create", "POST", "/apps", controllers.AppCreate)
	server.Route("app.delete", "DELETE", "/apps/{name}", controllers.AppDelete)
	server.Route("app.get", "GET", "/apps/{name}", controllers.AppGet)
	server.Route("app.list", "GET", "/apps", controllers.AppList)

	server.Route("build.create", "POST", "/apps/{app}/builds", controllers.BuildCreate)
	server.Route("build.get", "GET", "/apps/{app}/builds/{id}", controllers.BuildGet)
	server.Route("build.logs", "GET", "/apps/{app}/builds/{id}/logs", controllers.BuildLogs)
	server.Route("build.update", "PUT", "/apps/{app}/builds/{id}", controllers.BuildUpdate)

	server.Route("files.delete", "DELETE", "/apps/{app}/processes/{process}/files", controllers.FilesDelete)
	server.Route("files.upload", "POST", "/apps/{app}/processes/{process}/files", controllers.FilesUpload)

	server.Route("object.fetch", "GET", "/apps/{app}/objects/{key:.*}", controllers.ObjectFetch)
	server.Route("object.store", "POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	server.Route("process.list", "GET", "/apps/{app}/processes", controllers.ProcessList)
	server.Route("process.run", "POST", "/apps/{app}/processes", controllers.ProcessRun)
	server.Route("process.stop", "DELETE", "/apps/{app}/processes/{pid}", controllers.ProcessStop)

	server.Route("proxy", "POST", "/apps/{app}/processes/{process}/proxy/{port}", controllers.Proxy)

	server.Route("queue.fetch", "GET", "/apps/{app}/queues/{queue}", controllers.QueueFetch)
	server.Route("queue.store", "POST", "/apps/{app}/queues/{queue}", controllers.QueueStore)

	server.Route("release.create", "POST", "/apps/{app}/releases", controllers.ReleaseCreate)
	server.Route("release.get", "GET", "/apps/{app}/releases/{id}", controllers.ReleaseGet)

	server.Route("system.get", "GET", "/system", controllers.SystemGet)

	server.Route("table.fetch", "GET", "/apps/{app}/tables/{table}/indexes/{index}/{key}", controllers.TableFetch)
	server.Route("table.fetch.batch", "POST", "/apps/{app}/tables/{table}/indexes/{index}/batch", controllers.TableFetchBatch)
	server.Route("table.get", "GET", "/apps/{app}/tables/{table}", controllers.TableGet)
	server.Route("table.store", "POST", "/apps/{app}/tables/{table}", controllers.TableStore)
	server.Route("table.truncate", "POST", "/apps/{app}/tables/{table}/truncate", controllers.TableTruncate)
}
