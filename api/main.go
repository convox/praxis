package main

import (
	"net/http"
	"os"

	"github.com/convox/logger"
	"github.com/convox/nlogger"
	"github.com/convox/praxis/api/controllers"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func main() {
	if err := start(); err != nil {
		os.Exit(1)
	}
}

func start() error {
	log := logger.New("ns=api").At("start")

	r := mux.NewRouter()

	r.HandleFunc("/apps/{app}/tables", controllers.TableList)

	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(nlogger.New("ns=api", nil))
	n.UseHandler(r)

	port := "5000"

	log.Logf("port=%s", port)

	if err := http.ListenAndServe(":5000", n); err != nil {
		log.Error(err)
		return err
	}

	return nil
}
