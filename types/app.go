package types

import "time"

type App struct {
	Name string

	Release string
	Status  string
}

type Apps []App

type AppLogsOptions struct {
	Filter string
	Follow bool
	Since  time.Time
}
