package rack

type App struct {
	Name string
}

func (c *Client) AppCreate(name string) (app *App, err error) {
	err = c.Post("/apps", Params{"name": name}, &app)
	return
}
