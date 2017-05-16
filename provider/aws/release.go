package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
)

func (p *Provider) ReleaseCreate(app string, opts types.ReleaseCreateOptions) (*types.Release, error) {
	r, err := p.releaseFork(app)
	if err != nil {
		return nil, err
	}

	if opts.Build != "" {
		r.Build = opts.Build
	}

	if opts.Env != nil {
		r.Env = opts.Env
	}

	r.Stage = opts.Stage

	if err := p.releaseStore(r); err != nil {
		return nil, err
	}

	return r, nil
}

func (p *Provider) ReleaseGet(app, id string) (release *types.Release, err error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

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

	r, err := p.releaseFromAttributes(id, res.Attributes)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (p *Provider) ReleaseList(app string, opts types.ReleaseListOptions) (types.Releases, error) {
	if _, err := p.AppGet(app); err != nil {
		return nil, err
	}

	domain, err := p.appResource(app, "Releases")
	if err != nil {
		return nil, err
	}

	limit := coalescei(opts.Count, 10)

	req := &simpledb.SelectInput{
		ConsistentRead:   aws.Bool(true),
		SelectExpression: aws.String(fmt.Sprintf("select * from `%s` where created is not null order by created desc limit %d", domain, limit)),
	}

	releases := types.Releases{}

	res, err := p.SimpleDB().Select(req)
	if err != nil {
		return nil, err
	}

	for _, item := range res.Items {
		release, err := p.releaseFromAttributes(*item.Name, item.Attributes)
		if err != nil {
			return nil, err
		}

		releases = append(releases, *release)
	}

	return releases, nil
}

func (p *Provider) ReleaseLogs(app, id string, opts types.LogsOptions) (io.ReadCloser, error) {
	group, err := p.appResource(app, "Logs")
	if err != nil {
		return nil, err
	}

	stream := fmt.Sprintf("convox/release/%s", id)

	r, w := io.Pipe()

	go p.subscribeLogsCallback(group, stream, opts, w, func() bool {
		r, err := p.ReleaseGet(app, id)
		if err != nil {
			return false
		}

		switch r.Status {
		case "promoted", "failed":
			return false
		}

		return true
	})

	return r, nil
}

func (p *Provider) ReleasePromote(app string, id string) error {
	a, err := p.AppGet(app)
	if err != nil {
		return err
	}

	m, r, err := helpers.ReleaseManifest(p, app, id)
	if err != nil {
		return err
	}

	if r.Build == "" {
		return fmt.Errorf("no build for release: %s", r.Id)
	}

	topic, err := p.rackResource("NotificationTopic")
	if err != nil {
		return err
	}

	tp := map[string]interface{}{
		"App":      a,
		"Env":      r.Env,
		"Manifest": m,
		"Release":  r,
		"Version":  p.Version,
	}

	data, err := formationTemplate("app", tp)
	if err != nil {
		return err
	}

	domain, err := p.rackOutput("Domain")
	if err != nil {
		return err
	}

	updates := map[string]string{
		"Domain":   strings.ToLower(domain),
		"Password": p.Password,
		"Release":  r.Id,
	}

	stack := fmt.Sprintf("%s-%s", p.Name, app)

	params, err := p.cloudformationUpdateParameters(stack, data, updates)
	if err != nil {
		return err
	}

	rs, err := types.Key(4)
	if err != nil {
		return err
	}

	_, err = p.CloudFormation().UpdateStack(&cloudformation.UpdateStackInput{
		Capabilities:       []*string{aws.String("CAPABILITY_IAM")},
		ClientRequestToken: aws.String(fmt.Sprintf("%s-%s", r.Id, rs)),
		Parameters:         params,
		NotificationARNs:   []*string{aws.String(topic)},
		StackName:          aws.String(stack),
		TemplateBody:       aws.String(string(data)),
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) releaseFromAttributes(id string, attrs []*simpledb.Attribute) (*types.Release, error) {
	release := &types.Release{
		Env: types.Environment{},
		Id:  id,
	}

	// get app first so we can use it later
	for _, attr := range attrs {
		if *attr.Name == "app" {
			release.App = *attr.Value
		}
	}

	var err error

	for _, attr := range attrs {
		switch *attr.Name {
		case "app":
			release.App = *attr.Value
		case "build":
			release.Build = *attr.Value
		case "created":
			release.Created, err = time.Parse(helpers.SortableTime, *attr.Value)
			if err != nil {
				return nil, err
			}
		case "env":
			key := *attr.Value

			if key != "" {
				r, err := p.ObjectFetch(release.App, key)
				if err != nil {
					return nil, err
				}

				data, err := ioutil.ReadAll(r)
				if err != nil {
					return nil, err
				}

				if err := json.Unmarshal(data, &release.Env); err != nil {
					return nil, err
				}
			}
		case "stage":
			s, err := strconv.Atoi(*attr.Value)
			if err != nil {
				return nil, err
			}
			release.Stage = s
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

	rs, err := p.ReleaseList(app, types.ReleaseListOptions{Count: 1})
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

	attrs := []*simpledb.ReplaceableAttribute{
		{Replace: aws.Bool(true), Name: aws.String("app"), Value: aws.String(release.App)},
		{Replace: aws.Bool(true), Name: aws.String("build"), Value: aws.String(release.Build)},
		{Replace: aws.Bool(true), Name: aws.String("created"), Value: aws.String(release.Created.Format(helpers.SortableTime))},
		{Replace: aws.Bool(true), Name: aws.String("stage"), Value: aws.String(strconv.Itoa(release.Stage))},
		{Replace: aws.Bool(true), Name: aws.String("status"), Value: aws.String(release.Status)},
	}

	if len(release.Env) > 0 {
		data, err := json.Marshal(release.Env)
		if err != nil {
			return err
		}

		eo, err := p.ObjectStore(release.App, fmt.Sprintf("convox/release/%s/env", release.Id), bytes.NewReader(data), types.ObjectStoreOptions{})
		if err != nil {
			return err
		}

		attrs = append(attrs, &simpledb.ReplaceableAttribute{
			Name:  aws.String("env"),
			Value: aws.String(eo.Key),
		})
	}

	_, err = p.SimpleDB().PutAttributes(&simpledb.PutAttributesInput{
		Attributes: attrs,
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
