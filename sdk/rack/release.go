package rack

type Release struct {
	Id string
}

type ReleaseCreateOptions struct {
	Build string
	Env   map[string]string
}

func (c *Client) ReleaseCreate(app string, opts ReleaseCreateOptions) (*Release, error) {
	return &Release{Id: "R1234"}, nil
}

func (c *Client) ReleaseManifest(app string, id string) ([]byte, error) {
	return []byte{}, nil
}
