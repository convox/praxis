package aws

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/types"
	"github.com/fatih/color"
)

const (
	RackFormation = "https://s3.amazonaws.com/praxis-releases/release/%s/formation/rack.json"
)

func (p *Provider) SystemGet() (*types.System, error) {
	aid, err := p.accountID()
	if err != nil {
		return nil, err
	}

	stack, err := p.describeStack(p.Name)
	if err != nil {
		return nil, err
	}

	system := &types.System{
		Account: aid,
		Image:   fmt.Sprintf("convox/praxis:%s", p.Version),
		Name:    p.Name,
		Region:  os.Getenv("AWS_REGION"),
		Status:  humanStatus(*stack.StackStatus),
		Version: p.Version,
	}

	return system, nil
}

func (p *Provider) SystemInstall(name string, opts types.SystemInstallOptions) (string, error) {
	if opts.Version == "" {
		return "", fmt.Errorf("must specify a version to install")
	}

	template := fmt.Sprintf(RackFormation, opts.Version)

	_, err := p.CloudFormation().CreateStack(&cloudformation.CreateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		Parameters: []*cloudformation.Parameter{
			&cloudformation.Parameter{ParameterKey: aws.String("Password"), ParameterValue: aws.String(opts.Password)},
		},
		StackName: aws.String(name),
		Tags: []*cloudformation.Tag{
			{Key: aws.String("Name"), Value: aws.String(name)},
			{Key: aws.String("System"), Value: aws.String("convox")},
			{Key: aws.String("Type"), Value: aws.String("rack")},
			{Key: aws.String("Version"), Value: aws.String(opts.Version)},
		},
		TemplateURL: aws.String(template),
	})
	if err != nil {
		return "", err
	}

	if err := p.cloudformationProgress(name, opts); err != nil {
		return "", err
	}

	stack, err := p.describeStack(name)
	if err != nil {
		return "", err
	}
	if *stack.StackStatus == "ROLLBACK_COMPLETE" {
		return "", fmt.Errorf("installation failed")
	}

	return p.stackOutput(name, "Endpoint")
}

func (p *Provider) SystemLogs(opts types.LogsOptions) (io.ReadCloser, error) {
	group, err := p.rackResource("RackLogs")
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	go p.subscribeLogs(group, "", opts, w)

	return r, nil
}

func (p *Provider) SystemOptions() (map[string]string, error) {
	options := map[string]string{
		"streaming": "websocket",
	}

	return options, nil
}

func (p *Provider) SystemProxy(host string, port int, in io.Reader) (io.ReadCloser, error) {
	cn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	r, w := io.Pipe()

	go helpers.StreamAsync(cn, in, nil)
	go helpers.StreamAsync(w, cn, nil)

	return r, nil
}

func (p *Provider) SystemUninstall(name string, opts types.SystemInstallOptions) error {
	_, err := p.CloudFormation().DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	})
	if err != nil {
		return err
	}

	if err := p.cloudformationProgress(name, opts); err != nil {
		return err
	}

	return nil
}

func (p *Provider) SystemUpdate(opts types.SystemUpdateOptions) error {
	version, err := p.stackOutput(p.Name, "Version")
	if err != nil {
		return err
	}

	template := fmt.Sprintf(RackFormation, version)
	updates := map[string]string{}

	if opts.Version != "" {
		template = fmt.Sprintf(RackFormation, opts.Version)
	}

	if opts.Password != "" {
		updates["Password"] = opts.Password
	}

	res, err := http.Get(template)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	params, err := p.cloudformationUpdateParameters(p.Name, data, updates)
	if err != nil {
		return err
	}

	_, err = p.CloudFormation().UpdateStack(&cloudformation.UpdateStackInput{
		Capabilities: []*string{aws.String("CAPABILITY_IAM")},
		Parameters:   params,
		StackName:    aws.String(p.Name),
		TemplateURL:  aws.String(string(template)),
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) cloudformationProgress(name string, opts types.SystemInstallOptions) error {
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
			return nil // stack is gone, we're done
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

		stack, err := p.describeStack(name)
		if err != nil {
			return err
		}

		switch *stack.StackStatus {
		case "CREATE_COMPLETE":
			return nil
		case "ROLLBACK_COMPLETE":
			return fmt.Errorf("installation failed")
		}

		time.Sleep(2 * time.Second)
	}
}
