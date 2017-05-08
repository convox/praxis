package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/convox/praxis/helpers"
)

type queueHandler func(body string) error

func (p *Provider) workers() {
	go helpers.PrintError(p.autoscale())
	go helpers.PrintError(p.workerQueues())

	select {}
}

func (p *Provider) autoscale() error {
	asg, err := p.rackResource("Instances")
	if err != nil {
		return err
	}

	for {
		time.Sleep(5 * time.Second)

		hres, err := p.AutoScaling().DescribeScalingActivities(&autoscaling.DescribeScalingActivitiesInput{
			AutoScalingGroupName: aws.String(asg),
		})
		if err != nil {
			fmt.Printf("ns=provider.aws at=autoscale error=%q\n", err)
			continue
		}
		if len(hres.Activities) < 1 {
			fmt.Printf("ns=provider.aws at=autoscale error=%q\n", "no activities")
			continue
		}
		if hres.Activities[0].EndTime == nil {
			fmt.Printf("ns=provider.aws at=autoscale error=%q\n", "in progress")
			continue
		}

		scaled := *hres.Activities[0].EndTime

		cis, err := p.containerInstances()
		if err != nil {
			fmt.Printf("ns=provider.aws at=autoscale error=%q\n", err)
			continue
		}

		current := int64(len(cis))
		desired := current

		ss, err := p.clusterServices()
		if err != nil {
			fmt.Printf("ns=provider.aws at=autoscale error=%q\n", err)
			continue
		}

		if len(cis) < 1 {
			fmt.Printf("ns=provider.aws at=autoscale error=%q\n", "no instances")
			continue
		}

		available := map[string]int64{}
		scheduled := map[string]int64{}
		single := map[string]int64{}

		for i, ci := range cis {
			for _, r := range ci.RegisteredResources {
				switch *r.Type {
				case "INTEGER":
					if i == 0 {
						single[*r.Name] = *r.IntegerValue
					}
					available[*r.Name] += *r.IntegerValue
				}
			}
		}

		for _, s := range ss {
			for _, d := range s.Deployments {
				td, err := p.fetchTaskDefinition(*d.TaskDefinition)
				if err != nil {
					fmt.Printf("ns=provider.aws at=autoscale error=%q\n", err)
					continue
				}

				for _, cd := range td.ContainerDefinitions {
					if d.DesiredCount != nil && cd.Cpu != nil && cd.Memory != nil {
						scheduled["CPU"] += (*d.DesiredCount * *cd.Cpu)
						scheduled["MEMORY"] += (*d.DesiredCount * *cd.Memory)
					}
				}
			}
		}

		needed := int64(0)
		extra := int64(current)

		for m := range scheduled {
			ce := ((available[m] - scheduled[m]) / single[m])
			cs := (scheduled[m] / single[m]) + 1

			if ce < extra {
				extra = ce
			}

			if cs > needed {
				needed = cs
			}
		}

		if extra >= 2 {
			desired = current - 1
		}

		if needed > desired {
			desired = needed
		}

		bump := false

		for _, s := range ss {
			for _, e := range s.Events {
				if strings.Index(*e.Message, "steady state") > -1 {
					fmt.Printf("ns=provider.aws at=autoscale service=%s state=steady\n", *s.ServiceName)
					break
				}

				if strings.Index(*e.Message, "has insufficient") > -1 && e.CreatedAt.Before(scaled) {
					fmt.Printf("ns=provider.aws at=autoscale service=%s state=insufficient\n", *s.ServiceName)
					bump = true
					break
				}
			}
		}

		if desired <= current && bump {
			desired = current + 1
		}

		if desired < 2 {
			desired = 2
		}

		fmt.Printf("ns=provider.aws at=autoscale current=%d desired=%d\n", current, desired)

		if desired != current {
			_, err := p.AutoScaling().SetDesiredCapacity(&autoscaling.SetDesiredCapacityInput{
				AutoScalingGroupName: aws.String(asg),
				DesiredCapacity:      aws.Int64(desired),
				HonorCooldown:        aws.Bool(true),
			})
			if awsError(err) == "ScalingActivityInProgress" {
				fmt.Printf("ns=provider.aws at=autoscale scale=%d status=cooldown\n", desired)
				continue
			}
			if err != nil {
				fmt.Printf("ns=provider.aws at=autoscale scale=%d error=%q\n", desired, err)
				continue
			}

			fmt.Printf("ns=provider.aws at=autoscale scale=%d status=success\n", desired)
		}
	}
}

func (p *Provider) workerQueues() error {
	queue, err := p.rackResource("NotificationQueue")
	if err != nil {
		return err
	}

	return p.subscribeQueue(queue, p.handleNotifications)
}

func (p *Provider) handleNotifications(body string) error {
	var item struct {
		Message   string
		Timestamp time.Time
	}

	if err := json.Unmarshal([]byte(body), &item); err != nil {
		return err
	}

	msg := map[string]string{}

	for _, line := range strings.Split(item.Message, "\n") {
		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			msg[parts[0]] = strings.Trim(parts[1], "'")
		}
	}

	if msg["ClientRequestToken"] == "null" {
		return nil
	}

	group, err := p.stackResource(msg["StackName"], "Logs")
	if err != nil {
		return err
	}

	stream := fmt.Sprintf("convox/release/%s", msg["ClientRequestToken"])

	app := strings.TrimPrefix(msg["StackName"], fmt.Sprintf("%s-", p.Name))

	p.writeLogf(group, stream, "%-20s  %-28s  %s", msg["ResourceStatus"], msg["LogicalResourceId"], msg["ResourceType"])

	if msg["LogicalResourceId"] == msg["StackName"] {
		r, err := p.ReleaseGet(app, msg["ClientRequestToken"])
		if err != nil {
			return err
		}

		switch msg["ResourceStatus"] {
		case "UPDATE_IN_PROGRESS":
			r.Status = "running"
		case "UPDATE_COMPLETE":
			p.writeLogf(group, stream, "release promoted: %s", msg["ClientRequestToken"])
			r.Status = "complete"
		}

		if err := p.releaseStore(r); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) subscribeQueue(queue string, fn queueHandler) error {
	for {
		res, err := p.SQS().ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(queue),
			AttributeNames:        []*string{aws.String("All")},
			MessageAttributeNames: []*string{aws.String("All")},
			MaxNumberOfMessages:   aws.Int64(10),
			VisibilityTimeout:     aws.Int64(20),
			WaitTimeSeconds:       aws.Int64(10),
		})
		if err != nil {
			return err
		}

		for _, m := range res.Messages {
			if err := fn(*m.Body); err != nil {
				fmt.Fprintf(os.Stderr, "processQueue %s handler error: %s\n", queue, err)
			}

			_, err := p.SQS().DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queue),
				ReceiptHandle: m.ReceiptHandle,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "processQueue DeleteMessage error: %s\n", err)
			}
		}
	}
}
