package aws

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
)

func (p *Provider) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error) {
	a, err := p.AppGet(app)
	if err != nil {
		return nil, err
	}

	r, err := p.releaseFork(app)
	if err != nil {
		return nil, err
	}

	if opts.Build != "" {
		r.Build = opts.Build
	}

	if len(opts.Env) > 0 {
		r.Env = opts.Env
	}

	if err := p.releaseStore(r); err != nil {
		return nil, err
	}

	if r.Build == "" {
		return r, nil
	}

	b, err := p.BuildGet(app, r.Build)
	if err != nil {
		return nil, err
	}

	m, err := manifest.Load([]byte(b.Manifest))
	if err != nil {
		return nil, err
	}

	tp := map[string]interface{}{
		"App":      a,
		"Env":      r.Env,
		"Manifest": m,
		"Release":  r,
	}

	data, err := formationTemplate("app", tp)
	if err != nil {
		return nil, err
	}

	domain, err := p.rackOutput("Domain")
	if err != nil {
		return nil, err
	}

	updates := map[string]string{
		"Domain":  strings.ToLower(domain),
		"Release": r.Id,
	}

	stack := fmt.Sprintf("%s-%s", p.Name, app)

	res, err := p.CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stack),
	})
	if err != nil {
		return nil, err
	}
	if len(res.Stacks) != 1 {
		return nil, fmt.Errorf("could not find stack: %s", stack)
	}

	params := []*cloudformation.Parameter{}

	for _, p := range res.Stacks[0].Parameters {
		if u, ok := updates[*p.ParameterKey]; ok {
			params = append(params, &cloudformation.Parameter{
				ParameterKey:   p.ParameterKey,
				ParameterValue: aws.String(u),
			})
			delete(updates, *p.ParameterKey)
		} else {
			params = append(params, &cloudformation.Parameter{
				ParameterKey:     p.ParameterKey,
				UsePreviousValue: aws.Bool(true),
			})
		}
	}

	for k, v := range updates {
		params = append(params, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(v),
		})
	}

	np, err := formationParameters(data)
	if err != nil {
		return nil, err
	}

	for i, q := range params {
		found := false
		for _, p := range np {
			if p == *q.ParameterKey {
				found = true
				break
			}
		}
		if !found {
			params = append(params[:i], params[i+1:]...)
		}
	}

	// fmt.Printf("string(data) = %+v\n", string(data))

	_, err = p.CloudFormation().UpdateStack(&cloudformation.UpdateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		Parameters:   params,
		StackName:    aws.String(fmt.Sprintf("%s-%s", p.Name, app)),
		TemplateBody: aws.String(string(data)),
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (p *Provider) ReleaseGet(app, id string) (release *types.Release, err error) {
	domain, err := p.appResource(app, "Releases")
	if err != nil {
		return nil, err
	}

	res, err := p.SimpleDB().GetAttributes(&simpledb.GetAttributesInput{
		DomainName: aws.String(domain),
		ItemName:   aws.String(id),
	})
	if err != nil {
		return nil, err
	}

	return releaseFromAttributes(id, res.Attributes)
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
			if release, err := releaseFromAttributes(*item.Name, item.Attributes); err == nil {
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

func (p *Provider) ReleaseLogs(app, id string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("unimplemented")
}

func releaseFromAttributes(id string, attrs []*simpledb.Attribute) (*types.Release, error) {
	release := &types.Release{
		Id: id,
	}

	var err error

	for _, attr := range attrs {
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
		case "status":
			release.Status = *attr.Value
		}
	}

	return release, nil
}

func (p *Provider) releaseFork(app string) (*types.Release, error) {
	r := &types.Release{
		Id:      types.Id("R", 10),
		App:     app,
		Status:  "created",
		Created: time.Now().UTC(),
	}

	rs, err := p.ReleaseList(app)
	if err != nil {
		return nil, err
	}

	if len(rs) > 0 {
		r.Build = rs[0].Build
		r.Env = rs[0].Env
	}

	return r, nil
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
			{Name: aws.String("status"), Value: aws.String(release.Status)},
		},
		DomainName: aws.String(domain),
		ItemName:   aws.String(release.Id),
	})

	return err
}

func formationParameters(data []byte) ([]string, error) {
	var t struct {
		Parameters map[string]interface{}
	}

	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	params := []string{}

	for k := range t.Parameters {
		params = append(params, k)
	}

	return params, nil
}
