package aws

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/convox/praxis/types"
)

func (p *Provider) SystemGet() (*types.System, error) {
	system := &types.System{
		Name:    p.Rack,
		Image:   fmt.Sprintf("convox/praxis:%s", os.Getenv("VERSION")),
		Version: os.Getenv("VERSION"),
	}

	return system, nil
}

func (p *Provider) SystemImage() (string, error) {
	res, err := p.CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(p.Rack),
	})
	if err != nil {
		return "", err
	}

	fmt.Printf("res = %+v\n", res)

	return fmt.Sprintf("convox/praxis:%s", ""), nil
}
