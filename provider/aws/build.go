package aws

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		Created: time.Now().UTC(),
	}

	if err := p.buildStore(build); err != nil {
		return nil, err
	}

	ar, err := p.AppRegistry(app)
	if err != nil {
		return nil, err
	}

	repo, err := p.appResource(app, "Repository")
	if err != nil {
		return nil, err
	}

	sys, err := p.SystemGet()
	if err != nil {
		return nil, err
	}

	pid, err := p.ProcessStart(app, types.ProcessRunOptions{
		Command: fmt.Sprintf("build -id %s -url %s", id, url),
		Environment: map[string]string{
			"BUILD_APP":    app,
			"BUILD_PREFIX": fmt.Sprintf("%s-%s", p.Name, app),
			"BUILD_PUSH":   fmt.Sprintf("%s/%s", ar.Hostname, repo),
		},
		Name:    fmt.Sprintf("%s-%s-build-%s", p.Name, app, id),
		Image:   sys.Image,
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

	if id == "" {
		return nil, fmt.Errorf("blank id")
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

	return p.buildFromAttributes(id, res.Attributes)
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
		build, err := p.buildFromAttributes(*item.Name, item.Attributes)
		if err != nil {
			return nil, err
		}

		builds[i] = *build
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

func (p *Provider) buildFromAttributes(id string, attrs []*simpledb.Attribute) (*types.Build, error) {
	build := &types.Build{Id: id}

	var err error

	// get app first so we can use it later
	for _, attr := range attrs {
		if *attr.Name == "app" {
			build.App = *attr.Value
		}
	}

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
			key := *attr.Value

			if key != "" {
				r, err := p.ObjectFetch(build.App, key)
				if err != nil {
					return nil, err
				}

				data, err := ioutil.ReadAll(r)
				if err != nil {
					return nil, err
				}

				build.Manifest = string(data)
			}
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

	attrs := []*simpledb.ReplaceableAttribute{
		{Replace: aws.Bool(true), Name: aws.String("app"), Value: aws.String(build.App)},
		{Replace: aws.Bool(true), Name: aws.String("created"), Value: aws.String(build.Created.Format(sortableTime))},
		{Replace: aws.Bool(true), Name: aws.String("ended"), Value: aws.String(build.Ended.Format(sortableTime))},
		{Replace: aws.Bool(true), Name: aws.String("process"), Value: aws.String(build.Process)},
		{Replace: aws.Bool(true), Name: aws.String("release"), Value: aws.String(build.Release)},
		{Replace: aws.Bool(true), Name: aws.String("started"), Value: aws.String(build.Started.Format(sortableTime))},
		{Replace: aws.Bool(true), Name: aws.String("status"), Value: aws.String(build.Status)},
	}

	if build.Manifest != "" {
		mo, err := p.ObjectStore(build.App, fmt.Sprintf("convox/build/%s/manifest", build.Id), bytes.NewReader([]byte(build.Manifest)), types.ObjectStoreOptions{})
		if err != nil {
			return err
		}

		attrs = append(attrs, &simpledb.ReplaceableAttribute{
			Name:  aws.String("manifest"),
			Value: aws.String(mo.Key),
		})
	}

	_, err = p.SimpleDB().PutAttributes(&simpledb.PutAttributesInput{
		Attributes: attrs,
		DomainName: aws.String(domain),
		ItemName:   aws.String(build.Id),
	})

	return err
}
