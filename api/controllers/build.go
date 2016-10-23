package controllers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/convox/praxis/provider"
	"github.com/gorilla/mux"
)

func BuildCreate(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	opts := provider.BuildCreateOptions{
		Cache: true,
	}

	build, err := Provider.BuildCreate(vars["app"], r.FormValue("url"), opts)
	fmt.Printf("build = %+v\n", build)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return err
	}

	env, err := Provider.EnvironmentLoad(vars["app"])
	fmt.Printf("env = %+v\n", env)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return err
	}

	release, err := Provider.ReleaseCreate(vars["app"], build, env)
	fmt.Printf("release = %+v\n", release)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return err
	}

	build.Release = release.Id

	fmt.Printf("build = %+v\n", build)

	if err := Provider.BuildSave(build); err != nil {
		return err
	}

	return Render(w, build)
}

func BuildGet(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	build, err := Provider.BuildLoad(vars["app"], vars["build"])
	if err != nil {
		return err
	}

	return Render(w, build)
}

func BuildLogs(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	rd, err := Provider.BuildLogs(vars["app"], vars["build"])
	if err != nil {
		return err
	}

	if _, err := io.Copy(flushWriter{w}, rd); err != nil {
		return err
	}

	return nil
}
