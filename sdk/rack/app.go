package rack

import "fmt"

type App struct {
	Name string
}

func (c *Client) AppCreate(name string) (app *App, err error) {
	ro := RequestOptions{
		Params: Params{
			"name": name,
		},
	}

	err = c.Post("/apps", ro, &app)
	return
}

func (c *Client) AppDelete(name string) error {
	return c.Delete(fmt.Sprintf("/apps/%s", name), RequestOptions{}, nil)
}
