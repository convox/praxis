package rack

import (
	"fmt"

	"github.com/convox/praxis/types"
)

type Release types.Release

type ReleaseCreateOptions struct {
	Build string
	Env   map[string]string
}

func (c *Client) ReleaseCreate(app string, opts ReleaseCreateOptions) (*Release, error) {
	return &Release{Id: "R1234"}, nil
}

func (c *Client) ReleaseGet(app, id string) (release *Release, err error) {
	err = c.Get(fmt.Sprintf("/apps/%s/releases/%s", app, id), &release)
	return
}

func (c *Client) ReleaseManifest(app string, id string) ([]byte, error) {
	return []byte{}, nil
}
