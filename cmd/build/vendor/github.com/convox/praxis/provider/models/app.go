package models

type App struct {
	Name string `json:"name"`
}

type Apps []App

type AppCreateOptions struct {
}
