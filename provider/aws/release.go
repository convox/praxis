package aws

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/convox/praxis/types"
)

func (p *Provider) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error) {
	id := types.Id("R", 10)

	release := &types.Release{
		Id:      id,
		App:     app,
		Build:   opts.Build,
		Env:     opts.Env,
		Created: time.Now(),
	}

	if err := p.releaseStore(release); err != nil {
		return nil, err
	}

	return release, nil
}

func (p *Provider) ReleaseGet(app, id string) (release *types.Release, err error) {
	return nil, nil
}

func (p *Provider) ReleaseList(app string) (types.Releases, error) {
	domain, err := p.appResource(app, "Releases")
	if err != nil {
		return nil, err
	}

	req := &simpledb.SelectInput{
		ConsistentRead:   aws.Bool(true),
		SelectExpression: aws.String(fmt.Sprintf("select * from `%s` where created is not null order by created desc", domain)),
	}

	releases := types.Releases{}

	err = p.SimpleDB().SelectPages(req, func(res *simpledb.SelectOutput, last bool) bool {
		for _, item := range res.Items {
			if release, err := releaseFromItem(item); err == nil {
				releases = append(releases, *release)
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func (p *Provider) ReleasePromote(app, id string) error {
	return nil
}

func releaseFromItem(item *simpledb.Item) (*types.Release, error) {
	release := &types.Release{
		Id: *item.Name,
	}

	var err error

	for _, attr := range item.Attributes {
		switch *attr.Name {
		case "app":
			release.App = *attr.Value
		case "build":
			release.Build = *attr.Value
		case "created":
			release.Created, err = time.Parse(sortableTime, *attr.Value)
			if err != nil {
				return nil, err
			}
		case "env":
			if err := json.Unmarshal([]byte(*attr.Value), &release.Env); err != nil {
				return nil, err
			}
		}
	}

	return release, nil
}

func (p *Provider) releaseStore(release *types.Release) error {
	domain, err := p.appResource(release.App, "Releases")
	if err != nil {
		return err
	}

	env, err := json.Marshal(release.Env)
	if err != nil {
		return err
	}

	_, err = p.SimpleDB().PutAttributes(&simpledb.PutAttributesInput{
		Attributes: []*simpledb.ReplaceableAttribute{
			{Name: aws.String("app"), Value: aws.String(release.App)},
			{Name: aws.String("build"), Value: aws.String(release.Build)},
			{Name: aws.String("created"), Value: aws.String(release.Created.Format(sortableTime))},
			{Name: aws.String("env"), Value: aws.String(string(env))},
		},
		DomainName: aws.String(domain),
		ItemName:   aws.String(release.Id),
	})

	return err
}
