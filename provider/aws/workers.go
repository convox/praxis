package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type queueHandler func(body string) error

func (p *Provider) workers() {
	go func() {
		if err := p.workerQueues(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
	}()

	select {}
}

func (p *Provider) workerQueues() error {
	queue, err := p.rackResource("NotificationQueue")
	if err != nil {
		return err
	}

	err = p.subscribeQueue(queue, func(body string) error {
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
	})

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
