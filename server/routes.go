package server

import (
	"github.com/convox/api"
	"github.com/convox/praxis/server/controllers"
)

func Routes(api *api.Server) {
	api.Route("app.create", "POST", "/apps", controllers.AppCreate)
	api.Route("app.delete", "DELETE", "/apps/{name}", controllers.AppDelete)
	api.Route("app.get", "GET", "/apps/{name}", controllers.AppGet)
	api.Route("app.list", "GET", "/apps", controllers.AppList)

	api.Route("build.create", "POST", "/apps/{app}/builds", controllers.BuildCreate)
	api.Route("build.get", "GET", "/apps/{app}/builds/{id}", controllers.BuildGet)
	api.Route("build.logs", "GET", "/apps/{app}/builds/{id}/logs", controllers.BuildLogs)
	api.Route("build.update", "PUT", "/apps/{app}/builds/{id}", controllers.BuildUpdate)

	api.Route("files.delete", "DELETE", "/apps/{app}/processes/{process}/files", controllers.FilesDelete)
	api.Route("files.upload", "POST", "/apps/{app}/processes/{process}/files", controllers.FilesUpload)

	api.Route("object.fetch", "GET", "/apps/{app}/objects/{key:.*}", controllers.ObjectFetch)
	api.Route("object.store", "POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	api.Route("process.list", "GET", "/apps/{app}/processes", controllers.ProcessList)
	api.Route("process.run", "POST", "/apps/{app}/processes", controllers.ProcessRun)
	api.Route("process.stop", "DELETE", "/apps/{app}/processes/{pid}", controllers.ProcessStop)

	api.Route("proxy", "POST", "/apps/{app}/processes/{process}/proxy/{port}", controllers.Proxy)

	api.Route("queue.fetch", "GET", "/apps/{app}/queues/{queue}", controllers.QueueFetch)
	api.Route("queue.store", "POST", "/apps/{app}/queues/{queue}", controllers.QueueStore)

	api.Route("release.create", "POST", "/apps/{app}/releases", controllers.ReleaseCreate)
	api.Route("release.get", "GET", "/apps/{app}/releases/{id}", controllers.ReleaseGet)

	api.Route("system.get", "GET", "/system", controllers.SystemGet)

	api.Route("table.fetch", "GET", "/apps/{app}/tables/{table}/id/{id}", controllers.TableFetch)
	api.Route("table.fetchindex", "GET", "/apps/{app}/tables/{table}/{index}/{key}", controllers.TableFetchIndex)
	api.Route("table.get", "GET", "/apps/{app}/tables/{table}", controllers.TableGet)
	api.Route("table.store", "POST", "/apps/{app}/tables/{table}", controllers.TableStore)
}
