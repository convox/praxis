package rack

type Build struct {
	Id string
}

func (c *Client) BuildCreate(app string, url string) (*Build, error) {
	return &Build{Id: "B1234"}, nil
}
