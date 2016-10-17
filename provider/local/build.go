package local

import (
	"fmt"

	"github.com/convox/praxis/provider/models"
)

func (p *Provider) BuildCreate(app, url string, opts models.BuildCreateOptions) (*models.Build, error) {
	ps, err := p.ProcessRun(app, "api", models.ProcessRunOptions{
		Command: []string{"build", "-url", url},
	})
	fmt.Printf("ps = %+v\n", ps)
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return nil, err
	}

	fmt.Printf("ps = %+v\n", ps)

	return nil, nil
}
