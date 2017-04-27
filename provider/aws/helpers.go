package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func coalesce(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}

	return ""
}

func (p *Provider) cloudformationUpdateParameters(stack string, body []byte, updates map[string]string) ([]*cloudformation.Parameter, error) {
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

	np, err := formationParameters(body)
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

	return params, nil
}
