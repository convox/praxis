package rack

import "fmt"

type App struct {
	Name string
}

func (c *Client) AppCreate(name string) (app *App, err error) {
	err = c.Post("/apps", Params{"name": name}, &app)
	return
}

func (c *Client) AppDelete(name string) error {
	return c.Delete(fmt.Sprintf("/apps/%s", name), nil)
}
