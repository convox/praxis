package server

import (
	"github.com/convox/praxis/api"
	"github.com/convox/praxis/server/controllers"
)

type Server struct {
	*api.Server
}

func New() *Server {
	api := api.New("rack", "convox.rack")

	api.Route("app.create", "POST", "/apps", controllers.AppCreate)
	api.Route("app.delete", "DELETE", "/apps/{app}", controllers.AppDelete)
	api.Route("app.list", "GET", "/apps", controllers.AppList)

	api.Route("build.create", "POST", "/apps/{app}/builds", controllers.BuildCreate)
	api.Route("build.get", "GET", "/apps/{app}/builds/{id}", controllers.BuildGet)
	api.Route("build.logs", "GET", "/apps/{app}/builds/{id}/logs", controllers.BuildLogs)
	api.Route("build.update", "PUT", "/apps/{app}/builds/{id}", controllers.BuildUpdate)

	api.Route("object.fetch", "GET", "/apps/{app}/objects/{key:.*}", controllers.ObjectFetch)
	api.Route("object.store", "POST", "/apps/{app}/objects/{key:.*}", controllers.ObjectStore)

	api.Route("process.run", "POST", "/apps/{app}/processes", controllers.ProcessRun)

	api.Route("release.create", "POST", "/apps/{app}/releases", controllers.ReleaseCreate)
	api.Route("release.get", "GET", "/apps/{app}/releases/{id}", controllers.ReleaseGet)

	api.Route("system.get", "GET", "/system", controllers.SystemGet)

	return &Server{Server: api}
}
