package rack

type App struct {
	Name string
}

func (c *Client) AppCreate(name string) (*App, error) {
	return &App{Name: name}, nil
}
