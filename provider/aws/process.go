package aws

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/convox/praxis/types"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p *Provider) ProcessGet(app, pid string) (*types.Process, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) ProcessLogs(app, pid string) (io.ReadCloser, error) {
	r, w := io.Pipe()

	go p.cloudwatchLogStream(app, pid, w)

	return r, nil
}

func (p *Provider) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	return 0, fmt.Errorf("unimplemented")
}

func (p *Provider) ProcessStart(app string, opts types.ProcessRunOptions) (string, error) {
	fmt.Printf("opts = %+v\n", opts)

	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return "", err
	}

	td, err := p.taskDefinition(app, opts)
	if err != nil {
		return "", err
	}

	req := &ecs.RunTaskInput{
		Cluster:        aws.String(cluster),
		StartedBy:      aws.String(opts.Name),
		TaskDefinition: aws.String(td),
	}

	res, err := p.ECS().RunTask(req)
	if err != nil {
		return "", err
	}

	if len(res.Tasks) != 1 {
		return "", fmt.Errorf("unable to start process")
	}

	parts := strings.Split(*res.Tasks[0].TaskArn, "-")

	return parts[len(parts)-1], nil
}

func (p *Provider) ProcessStop(app, pid string) error {
	return fmt.Errorf("unimplemented")
}
