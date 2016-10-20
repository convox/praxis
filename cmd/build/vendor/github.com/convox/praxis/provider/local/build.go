package local

import (
	"fmt"

	"github.com/convox/praxis/provider/models"
)

func (p *Provider) BuildCreate(app, url string, opts models.BuildCreateOptions) (*models.Build, error) {
	ps, err := p.ProcessStart("system", "api", models.ProcessRunOptions{
		Command: []string{"build", "-url", url},
	})
	fmt.Printf("err = %+v\n", err)
	if err != nil {
		return nil, err
	}

	b := &models.Build{
		Id:      id("B"),
		Process: ps.Id,
		Status:  "running",
	}

	go p.waitProcess(app, b.Id, ps.Id)

	return b, nil
}

func (p *Provider) waitProcess(app, build, process string) error {
	attrs, err := p.TableLoad("system", "builds", build)
	if err != nil {
		return err
	}

	code, err := p.ProcessWait(app, process)
	if err != nil {
		return err
	}

	switch code {
	case 0:
		attrs["status"] = "complete"
	default:
		attrs["status"] = "error"
	}

	if err := p.TableSave("system", "builds", build, attrs); err != nil {
		return err
	}

	return nil
}

func (p *Provider) BuildLoad(app, id string) (*models.Build, error) {
	attrs, err := p.TableLoad("system", "builds", id)
	if err != nil {
		return nil, err
	}

	fmt.Printf("attrs = %+v\n", attrs)

	build := &models.Build{
		Id:      id,
		Process: attrs["process"],
		Status:  attrs["status"],
	}

	return build, nil
}

func (p *Provider) BuildSave(build *models.Build) error {
	attrs := map[string]string{
		"error":    build.Error,
		"ended":    build.Ended.Format(SortableTime),
		"logs":     build.Logs,
		"manifest": build.Manifest,
		"process":  fmt.Sprintf("%s", build.Process),
		"status":   build.Status,
	}

	if err := p.TableSave("system", "builds", build.Id, attrs); err != nil {
		return err
	}

	return nil
}
