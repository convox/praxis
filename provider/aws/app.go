package aws

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/convox/praxis/types"
)

func (p *Provider) AppCreate(name string) (*types.App, error) {
	data, err := formationTemplate("app", nil)
	if err != nil {
		return nil, err
	}

	_, err = p.CloudFormation().CreateStack(&cloudformation.CreateStackInput{
		Parameters: []*cloudformation.Parameter{
			{ParameterKey: aws.String("Password"), ParameterValue: aws.String(p.Password)},
			{ParameterKey: aws.String("Rack"), ParameterValue: aws.String(p.Name)},
		},
		StackName: aws.String(fmt.Sprintf("%s-%s", p.Name, name)),
		Tags: []*cloudformation.Tag{
			{Key: aws.String("Name"), Value: aws.String(name)},
			{Key: aws.String("Rack"), Value: aws.String(p.Name)},
			{Key: aws.String("System"), Value: aws.String("convox")},
			{Key: aws.String("Type"), Value: aws.String("app")},
			{Key: aws.String("Version"), Value: aws.String(p.Version)},
		},
		TemplateBody: aws.String(string(data)),
	})
	if awsError(err) == "AlreadyExistsException" {
		return nil, fmt.Errorf("app already exists: %s", name)
	}
	if err != nil {
		return nil, err
	}

	return p.AppGet(name)
}

func (p *Provider) AppDelete(name string) error {
	bucket, _ := p.appResource(name, "Bucket")

	_, err := p.CloudFormation().DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(fmt.Sprintf("%s-%s", p.Name, name)),
	})
	if err != nil {
		return err
	}

	if bucket != "" {
		if err := p.deleteBucket(bucket); err != nil {
			fmt.Printf("ns=provider.aws at=app.delete error=%q\n", err)
		}
	}

	return nil
}

func (p *Provider) AppGet(name string) (*types.App, error) {
	stack, err := p.describeStack(fmt.Sprintf("%s-%s", p.Name, name))
	if awsError(err) == "ValidationError" {
		return nil, fmt.Errorf("no such app: %s", name)
	}
	if err != nil {
		return nil, err
	}

	app := p.appFromStack(stack)

	if app == nil {
		return nil, fmt.Errorf("no such app: %s", name)
	}

	if app.Status == "creating" {
		return app, nil
	}

	return app, nil
}

func (p *Provider) AppList() (types.Apps, error) {
	req := &cloudformation.DescribeStacksInput{}

	apps := types.Apps{}

	err := p.CloudFormation().DescribeStacksPages(req, func(res *cloudformation.DescribeStacksOutput, last bool) bool {
		for _, stack := range res.Stacks {
			if app := p.appFromStack(stack); app != nil {
				apps = append(apps, *app)
			}
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (p *Provider) AppLogs(app string, opts types.LogsOptions) (io.ReadCloser, error) {
	group, err := p.appResource(app, "Logs")
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	go p.subscribeLogs(group, "", opts, w)

	return r, nil
}

func (p *Provider) AppRegistry(app string) (*types.Registry, error) {
	account, err := p.accountID()
	if err != nil {
		return nil, err
	}

	res, err := p.ECR().GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{
		RegistryIds: []*string{aws.String(account)},
	})

	if err != nil {
		return nil, err
	}
	if len(res.AuthorizationData) != 1 {
		return nil, fmt.Errorf("no authorization data")
	}

	token, err := base64.StdEncoding.DecodeString(*res.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(token), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid auth data")
	}

	registry := &types.Registry{
		Hostname: fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", account, p.Region),
		Username: parts[0],
		Password: parts[1],
	}

	return registry, nil
}

func (p *Provider) appFromStack(stack *cloudformation.Stack) *types.App {
	outputs := map[string]string{}
	params := map[string]string{}
	tags := map[string]string{}

	for _, o := range stack.Outputs {
		outputs[*o.OutputKey] = *o.OutputValue
	}

	for _, p := range stack.Parameters {
		params[*p.ParameterKey] = *p.ParameterValue
	}

	for _, t := range stack.Tags {
		tags[*t.Key] = *t.Value
	}

	if tags["System"] != "convox" || tags["Rack"] != p.Name || tags["Type"] != "app" {
		return nil
	}

	name, ok := tags["Name"]
	if !ok {
		return nil
	}

	return &types.App{
		Name:    name,
		Release: outputs["Release"],
		Status:  humanStatus(*stack.StackStatus),
	}
}
