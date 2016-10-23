package local

import (
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/convox/praxis/provider"
)

func (p *Provider) BuildCreate(app, url string, opts provider.BuildCreateOptions) (*provider.Build, error) {
	build := &provider.Build{
		Id:     id("B"),
		Status: "running",
	}

	if err := p.BuildSave(build); err != nil {
		return nil, err
	}

	auth := "{}"

	ps, err := p.ProcessStart("system", "api", provider.ProcessRunOptions{
		Command: []string{"build", "-method", "tgz", "-url", url},
		Environment: map[string]string{
			"BUILD_APP":      app,
			"BUILD_AUTH":     auth,
			"BUILD_ID":       build.Id,
			"BUILD_MANIFEST": opts.Manifest,
			"BUILD_PUSH":     "",
		},
	})
	if err != nil {
		return nil, err
	}

	build.Process = ps.Id

	if err := p.BuildSave(build); err != nil {
		return nil, err
	}

	go p.waitProcess(app, build.Id, ps.Id)

	return build, nil
}

func (p *Provider) waitProcess(app, build, process string) error {
	code, err := p.ProcessWait(app, process)
	if err != nil {
		return err
	}

	attrs, err := p.TableLoad("system", "builds", build)
	if err != nil {
		return err
	}

	switch code {
	case 0:
		attrs["status"] = "complete"
	default:
		attrs["error"] = fmt.Sprintf("exit: %d", code)
		attrs["status"] = "error"
	}

	return p.TableSave("system", "builds", build, attrs)
}

func (p *Provider) BuildLoad(app, id string) (*provider.Build, error) {
	attrs, err := p.TableLoad("system", "builds", id)
	if err != nil {
		return nil, err
	}

	build := &provider.Build{
		Error:    attrs["error"],
		Id:       id,
		Manifest: attrs["manifest"],
		Process:  attrs["process"],
		Release:  attrs["release"],
		Status:   attrs["status"],
	}

	if attrs["ended"] != "" {
		t, err := time.Parse(SortableTime, attrs["ended"])
		if err != nil {
			return nil, err
		}

		build.Ended = t
	}

	return build, nil
}

func (p *Provider) BuildLogs(app, id string) (io.ReadCloser, error) {
	build, err := p.BuildLoad(app, id)
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	cmd := exec.Command("docker", "logs", "-f", build.Process)

	cmd.Stdout = w
	cmd.Stderr = w

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		cmd.Wait()
		w.Close()
	}()

	return r, nil
}

func (p *Provider) BuildSave(build *provider.Build) error {
	attrs := map[string]string{
		"error":    build.Error,
		"ended":    build.Ended.Format(SortableTime),
		"logs":     build.Logs,
		"manifest": build.Manifest,
		"process":  fmt.Sprintf("%s", build.Process),
		"release":  build.Release,
		"status":   build.Status,
	}

	if err := p.TableSave("system", "builds", build.Id, attrs); err != nil {
		return err
	}

	return nil
}
