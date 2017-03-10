package server

import (
	"net/http"

	"github.com/convox/api"
	"github.com/convox/praxis/server/controllers"
)

func Routes(server *api.Server) {
	server.Route("root", "GET", "/", func(w http.ResponseWriter, r *http.Request, c *api.Context) error {
		w.Write([]byte("ok"))
		return nil
	})

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

	server.Route("key.decrypt", "POST", "/apps/{app}/keys/{key}/decrypt", controllers.KeyDecrypt)
	server.Route("key.encrypt", "POST", "/apps/{app}/keys/{key}/encrypt", controllers.KeyEncrypt)

	server.Route("object.fetch", "GET", "/apps/{app}/objects/{key:.*}", controllers.ObjectFetch)
	server.Route("object.store", "POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	server.Route("process.get", "GET", "/apps/{app}/processes/{pid}", controllers.ProcessGet)
	server.Route("process.logs", "GET", "/apps/{app}/processes/{pid}/logs", controllers.ProcessLogs)
	server.Route("process.list", "GET", "/apps/{app}/processes", controllers.ProcessList)
	server.Route("process.run", "POST", "/apps/{app}/processes/run", controllers.ProcessRun)
	server.Route("process.start", "POST", "/apps/{app}/processes/start", controllers.ProcessStart)
	server.Route("process.stop", "DELETE", "/apps/{app}/processes/{pid}", controllers.ProcessStop)

	server.Route("proxy", "POST", "/apps/{app}/processes/{process}/proxy/{port}", controllers.Proxy)

	server.Route("queue.fetch", "GET", "/apps/{app}/queues/{queue}", controllers.QueueFetch)
	server.Route("queue.store", "POST", "/apps/{app}/queues/{queue}", controllers.QueueStore)

	server.Route("release.create", "POST", "/apps/{app}/releases", controllers.ReleaseCreate)
	server.Route("release.get", "GET", "/apps/{app}/releases/{id}", controllers.ReleaseGet)
	server.Route("release.promote", "POST", "/apps/{app}/releases/{id}/promote", controllers.ReleasePromote)

	server.Route("system.get", "GET", "/system", controllers.SystemGet)

	server.Route("table.create", "POST", "/apps/{app}/tables/{table}", controllers.TableCreate)
	server.Route("table.get", "GET", "/apps/{app}/tables/{table}", controllers.TableGet)
	server.Route("table.list", "GET", "/apps/{app}/tables", controllers.TableList)
	server.Route("table.truncate", "POST", "/apps/{app}/tables/{table}/truncate", controllers.TableTruncate)
	server.Route("table.row.delete", "DELETE", "/apps/{app}/tables/{table}/indexes/{index}/{key}", controllers.TableRowDelete)
	server.Route("table.row.get", "GET", "/apps/{app}/tables/{table}/indexes/{index}/{key}", controllers.TableRowGet)
	server.Route("table.row.store", "POST", "/apps/{app}/tables/{table}/rows", controllers.TableRowStore)
	server.Route("table.rows.delete", "POST", "/apps/{app}/tables/{table}/indexes/{index}/batch/remove", controllers.TableRowsDelete)
	server.Route("table.rows.get", "POST", "/apps/{app}/tables/{table}/indexes/{index}/batch", controllers.TableRowsGet)
}
