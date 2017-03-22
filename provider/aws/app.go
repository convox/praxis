package aws

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/convox/praxis/types"
)

func (p *Provider) AppCreate(name string) (*types.App, error) {
	data, err := formationTemplate("app", nil)
	if err != nil {
		return nil, err
	}

	_, err = p.CloudFormation().CreateStack(&cloudformation.CreateStackInput{
		Parameters: []*cloudformation.Parameter{
			{ParameterKey: aws.String("Release"), ParameterValue: aws.String("")},
		},
		StackName: aws.String(fmt.Sprintf("%s-%s", p.Rack, name)),
		Tags: []*cloudformation.Tag{
			{Key: aws.String("Name"), Value: aws.String(name)},
			{Key: aws.String("Rack"), Value: aws.String(p.Rack)},
			{Key: aws.String("System"), Value: aws.String("convox")},
			{Key: aws.String("Type"), Value: aws.String("app")},
			{Key: aws.String("Version"), Value: aws.String("test")},
		},
		TemplateBody: aws.String(string(data)),
	})
	if err != nil {
		return nil, err
	}

	return p.AppGet(name)
}

func (p *Provider) AppDelete(name string) error {
	_, err := p.CloudFormation().DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(fmt.Sprintf("%s-%s", p.Rack, name)),
	})
	return err
}

func (p *Provider) AppGet(name string) (*types.App, error) {
	res, err := p.CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(fmt.Sprintf("%s-%s", p.Rack, name)),
	})
	if awsError(err) == "ValidationError" {
		return nil, fmt.Errorf("no such app: %s", name)
	}
	if err != nil {
		return nil, err
	}

	if len(res.Stacks) < 1 {
		return nil, fmt.Errorf("no such app: %s", name)
	}

	app := p.appFromStack(res.Stacks[0])

	if app == nil {
		return nil, fmt.Errorf("no such app: %s", name)
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

func (p *Provider) AppLogs(name string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) appFromStack(stack *cloudformation.Stack) *types.App {
	params := map[string]string{}
	tags := map[string]string{}

	for _, p := range stack.Parameters {
		params[*p.ParameterKey] = *p.ParameterValue
	}

	for _, t := range stack.Tags {
		tags[*t.Key] = *t.Value
	}

	if tags["System"] != "convox" || tags["Rack"] != p.Rack {
		return nil
	}

	name, ok := tags["Name"]
	if !ok {
		return nil
	}

	return &types.App{
		Name:    name,
		Release: params["Release"],
		Status:  appStatusFromStackStatus(*stack.StackStatus),
	}
}

func appStatusFromStackStatus(status string) string {
	switch status {
	case "CREATE_COMPLETE", "ROLLBACK_COMPLETE":
		return "running"
	case "CREATE_IN_PROGRESS":
		return "creating"
	case "DELETE_IN_PROGRESS":
		return "deleting"
	case "DELETE_FAILED":
		return "error"
	case "ROLLBACK_IN_PROGRESS":
		return "rollback"
	default:
		return status
	}
}
