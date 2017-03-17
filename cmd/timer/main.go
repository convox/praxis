package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/convox/praxis/sdk/rack"
	"github.com/convox/praxis/types"
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

	ch := make(chan error)

	c := cron.New()

	c.AddFunc(fmt.Sprintf("0 %s", flagSchedule), func() {
		opts := types.ProcessRunOptions{
			Command: flagCommand,
			Service: flagService,
		}

		pid, err := Rack.ProcessStart(flagApp, opts)
		if err != nil {
			ch <- err
			return
		}

		r, err := Rack.ProcessLogs(flagApp, pid)
		if err != nil {
			ch <- err
			return
		}

		fmt.Printf("timer running: %s\n", flagName)

		if _, err := io.Copy(os.Stdout, r); err != nil {
			ch <- err
			return
		}
	})

	c.Start()

	for err := range ch {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}

	return nil
}
