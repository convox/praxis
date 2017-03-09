package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/convox/praxis/sdk/rack"
	"github.com/robfig/cron"
)

var (
	Rack rack.Rack

	flagApp      string
	flagCommand  string
	flagName     string
	flagSchedule string
	flagService  string
)

func init() {
	r, err := rack.NewFromEnv()
	if err != nil {
		panic(err)
	}

	Rack = r
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.StringVar(&flagApp, "app", "", "app name")
	fs.StringVar(&flagCommand, "command", "", "timer command")
	fs.StringVar(&flagName, "name", "", "timer name")
	fs.StringVar(&flagSchedule, "schedule", "", "timer schedule")
	fs.StringVar(&flagService, "service", "", "timer service")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	c := cron.New()

	c.AddFunc(flagSchedule, func() {
		fmt.Printf("flagApp = %+v\n", flagApp)
		fmt.Printf("flagCommand = %+v\n", flagCommand)
		fmt.Printf("flagName = %+v\n", flagName)
		fmt.Printf("flagSchedule = %+v\n", flagSchedule)
		fmt.Printf("flagService = %+v\n", flagService)
	})

	return nil
}
