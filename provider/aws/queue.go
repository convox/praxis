package aws

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/convox/praxis/types"
)

func (p *Provider) QueueFetch(app, queue string, opts types.QueueFetchOptions) (map[string]string, error) {
	q, err := p.appResource(app, fmt.Sprintf("Queue%s", upperName(queue)))
	if err != nil {
		return nil, err
	}

	res, err := p.SQS().ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q),
		MaxNumberOfMessages: aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}
	if len(res.Messages) < 1 {
		return nil, nil
	}

	var body map[string]string

	if err := json.Unmarshal([]byte(*res.Messages[0].Body), &body); err != nil {
		return nil, err
	}

	_, err = p.SQS().DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q),
		ReceiptHandle: res.Messages[0].ReceiptHandle,
	})
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (p *Provider) QueueStore(app, queue string, attrs map[string]string) error {
	q, err := p.appResource(app, fmt.Sprintf("Queue%s", upperName(queue)))
	if err != nil {
		return err
	}

	data, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	_, err = p.SQS().SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(q),
		MessageBody: aws.String(string(data)),
	})
	if err != nil {
		return err
	}

	return nil
}
