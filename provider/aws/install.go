package aws

import (
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/convox/praxis/types"
	"github.com/fatih/color"
)

func (p *Provider) Install(name string, opts types.InstallOptions) (string, error) {
	version := coalesce(opts.Version, "latest")
	template := fmt.Sprintf("https://s3.amazonaws.com/praxis-releases/release/%s/formation/rack.json", version)

	key, err := types.Key(64)
	if err != nil {
		return "", err
	}

	_, err = p.CloudFormation().CreateStack(&cloudformation.CreateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		Parameters: []*cloudformation.Parameter{
			&cloudformation.Parameter{ParameterKey: aws.String("ApiKey"), ParameterValue: aws.String(key)},
			&cloudformation.Parameter{ParameterKey: aws.String("Version"), ParameterValue: aws.String(version)},
		},
		StackName:   aws.String(name),
		TemplateURL: aws.String(template),
	})
	if err != nil {
		return "", err
	}

	if err := p.installProgress(name, opts); err != nil {
		return "", err
	}

	return p.stackOutput(name, "Endpoint")
}

func (p *Provider) installProgress(name string, opts types.InstallOptions) error {
	w := opts.Output
	if w == nil {
		w = ioutil.Discard
	}

	events := map[string]cloudformation.StackEvent{}

	for {
		eres, err := p.CloudFormation().DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
			StackName: aws.String(name),
		})
		if err != nil {
			return fmt.Errorf("installation failed") // stack is gone, we're done
		}

		sort.Slice(eres.StackEvents, func(i, j int) bool { return eres.StackEvents[i].Timestamp.Before(*eres.StackEvents[j].Timestamp) })

		for _, e := range eres.StackEvents {
			if _, ok := events[*e.EventId]; !ok {
				line := fmt.Sprintf("%-20s  %-28s  %s", *e.ResourceStatus, *e.LogicalResourceId, *e.ResourceType)

				if !opts.Color {
					fmt.Fprintf(w, "%s\n", line)
				} else {
					switch *e.ResourceStatus {
					case "CREATE_IN_PROGRESS":
						fmt.Fprintf(w, "%s\n", color.YellowString(line))
					case "CREATE_COMPLETE":
						fmt.Fprintf(w, "%s\n", color.GreenString(line))
					case "CREATE_FAILED":
						fmt.Fprintf(w, "%s\n  ERROR: %s\n", color.RedString(line), *e.ResourceStatusReason)
					case "DELETE_IN_PROGRESS", "DELETE_COMPLETE", "ROLLBACK_IN_PROGRESS", "ROLLBACK_COMPLETE":
						fmt.Fprintf(w, "%s\n", color.RedString(line))
					default:
						fmt.Fprintf(w, "%s\n", line)
					}
				}

				events[*e.EventId] = *e
			}
		}

		sres, err := p.CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		})
		if err != nil {
			return fmt.Errorf("installation failed") // stack is gone, we're done
		}

		if sres == nil || len(sres.Stacks) < 1 {
			return fmt.Errorf("could not find stack: %s", name)
		}

		switch *sres.Stacks[0].StackStatus {
		case "CREATE_COMPLETE":
			return nil
		case "ROLLBACK_COMPLETE":
			return fmt.Errorf("installation failed")
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}
