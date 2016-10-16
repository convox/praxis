package client

type Build struct {
	Id string `json:"id"`
}

type BuildCreateOptions struct {
	Cache bool
}

func (c *Client) BuildCreate(url string, opts BuildCreateOptions) (*Build, error) {
	var build Build

	popts := PostOptions{
		Params: map[string]string{
			"url": url,
		},
	}

	if err := c.Post("/builds", &build, popts); err != nil {
		return nil, err
	}

	return nil, nil
}
