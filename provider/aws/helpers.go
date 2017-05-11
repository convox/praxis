package aws

import (
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

func coalescei(ints ...int) int {
	for _, i := range ints {
		if i > 0 {
			return i
		}
	}

	return 0
}

func (p *Provider) cloudformationUpdateParameters(stack string, body []byte, updates map[string]string) ([]*cloudformation.Parameter, error) {
	s, err := p.describeStack(stack)
	if err != nil {
		return nil, err
	}

	params := []*cloudformation.Parameter{}

	for _, p := range s.Parameters {
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
