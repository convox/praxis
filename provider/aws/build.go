package aws

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/convox/praxis/types"
)

func (p *Provider) BuildCreate(app, url string, opts types.BuildCreateOptions) (*types.Build, error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

	id := types.Id("B", 10)

	build := &types.Build{
		Id:      id,
		App:     app,
		Status:  "created",
		Created: time.Now(),
	}

	if err := p.buildStore(build); err != nil {
		return nil, err
	}

	registries, err := p.RegistryList()
	if err != nil {
		return nil, err
	}

	ar, err := p.appRegistry(app)
	if err != nil {
		return nil, err
	}

	repo, err := p.appResource(app, "Repository")
	if err != nil {
		return nil, err
	}

	auth, err := json.Marshal(append(registries, *ar))
	if err != nil {
		return nil, err
	}

	pid, err := p.ProcessStart(app, types.ProcessRunOptions{
		Command: fmt.Sprintf("build -id %s -url %s", id, url),
		Environment: map[string]string{
			"BUILD_APP":  app,
			"BUILD_AUTH": base64.StdEncoding.EncodeToString(auth),
			"BUILD_PUSH": fmt.Sprintf("%s/%s", ar.Hostname, repo),
		},
		Name:    fmt.Sprintf("%s-%s-build-%s", p.Rack, app, id),
		Image:   "convox/praxis:test8",
		Service: "build",
		Volumes: map[string]string{
			"/var/run/docker.sock": "/var/run/docker.sock",
		},
	})
	if err != nil {
		return nil, err
	}

	build.Process = pid

	if err := p.buildStore(build); err != nil {
		return nil, err
	}

	return build, nil
}

func (p *Provider) BuildGet(app, id string) (*types.Build, error) {
	domain, err := p.appResource(app, "Builds")
	if err != nil {
		return nil, err
	}

	req := &simpledb.GetAttributesInput{
		ConsistentRead: aws.Bool(true),
		DomainName:     aws.String(domain),
		ItemName:       aws.String(id),
	}

	res, err := p.SimpleDB().GetAttributes(req)
	if err != nil {
		return nil, err
	}

	return buildFromAttributes(id, res.Attributes)
}

func (p *Provider) BuildList(app string) (types.Builds, error) {
	domain, err := p.appResource(app, "Builds")
	if err != nil {
		return nil, err
	}

	req := &simpledb.SelectInput{
		ConsistentRead:   aws.Bool(true),
		SelectExpression: aws.String(fmt.Sprintf("select * from `%s` where created is not null order by created desc limit 10", domain)),
	}

	res, err := p.SimpleDB().Select(req)
	if err != nil {
		return nil, err
	}

	builds := make(types.Builds, len(res.Items))

	for i, item := range res.Items {
		if build, err := buildFromAttributes(*item.Name, item.Attributes); err == nil {
			builds[i] = *build
		}
	}

	return builds, nil
}

func (p *Provider) BuildLogs(app, id string) (io.ReadCloser, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	switch build.Status {
	case "running":
		return p.ProcessLogs(app, build.Process)
	default:
		return p.ObjectFetch(app, fmt.Sprintf("convox/builds/%s/log", id))
	}
}

func (p *Provider) BuildUpdate(app, id string, opts types.BuildUpdateOptions) (*types.Build, error) {
	build, err := p.BuildGet(app, id)
	if err != nil {
		return nil, err
	}

	if !opts.Ended.IsZero() {
		build.Ended = opts.Ended
	}

	if opts.Manifest != "" {
		build.Manifest = opts.Manifest
	}

	if opts.Release != "" {
		build.Release = opts.Release
	}

	if !opts.Started.IsZero() {
		build.Started = opts.Started
	}

	if opts.Status != "" {
		build.Status = opts.Status
	}

	if err := p.buildStore(build); err != nil {
		return nil, err
	}

	return nil, nil
}

func buildFromAttributes(id string, attrs []*simpledb.Attribute) (*types.Build, error) {
	build := &types.Build{Id: id}

	var err error

	for _, attr := range attrs {
		switch *attr.Name {
		case "app":
			build.App = *attr.Value
		case "created":
			build.Created, err = time.Parse(sortableTime, *attr.Value)
			if err != nil {
				return nil, err
			}
		case "ended":
			build.Ended, err = time.Parse(sortableTime, *attr.Value)
			if err != nil {
				return nil, err
			}
		case "manifest":
			build.Manifest = *attr.Value
		case "process":
			build.Process = *attr.Value
		case "release":
			build.Release = *attr.Value
		case "started":
			build.Started, err = time.Parse(sortableTime, *attr.Value)
			if err != nil {
				return nil, err
			}
		case "status":
			build.Status = *attr.Value
		}
	}

	return build, nil
}

func (p *Provider) buildStore(build *types.Build) error {
	domain, err := p.appResource(build.App, "Builds")
	if err != nil {
		return err
	}

	_, err = p.SimpleDB().PutAttributes(&simpledb.PutAttributesInput{
		Attributes: []*simpledb.ReplaceableAttribute{
			{Name: aws.String("app"), Value: aws.String(build.App)},
			{Name: aws.String("created"), Value: aws.String(build.Created.Format(sortableTime))},
			{Name: aws.String("ended"), Value: aws.String(build.Ended.Format(sortableTime))},
			{Name: aws.String("manifest"), Value: aws.String(build.Manifest)},
			{Name: aws.String("process"), Value: aws.String(build.Process)},
			{Name: aws.String("release"), Value: aws.String(build.Release)},
			{Name: aws.String("started"), Value: aws.String(build.Started.Format(sortableTime))},
			{Name: aws.String("status"), Value: aws.String(build.Status)},
		},
		DomainName: aws.String(domain),
		ItemName:   aws.String(build.Id),
	})

	return err
}
