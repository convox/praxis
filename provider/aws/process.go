package aws

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/convox/praxis/types"
	shellquote "github.com/kballard/go-shellquote"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p *Provider) ProcessGet(app, pid string) (*types.Process, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (p *Provider) ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

	pss := types.Processes{}

	req := &ecs.ListTasksInput{
		Cluster:    aws.String(cluster),
		MaxResults: aws.Int64(2),
	}

	for {
		res, err := p.ECS().ListTasks(req)
		if err != nil {
			return nil, err
		}

		dres, err := p.ECS().DescribeTasks(&ecs.DescribeTasksInput{
			Cluster: aws.String(cluster),
			Tasks:   res.TaskArns,
		})
		if err != nil {
			return nil, err
		}

		for _, t := range dres.Tasks {
			ps, err := p.processFromTask(app, t)
			if err != nil {
				return nil, err
			}

			if ps.App != app {
				continue
			}

			if opts.Service != "" && ps.Service != opts.Service {
				continue
			}

			pss = append(pss, *ps)
		}

		if res.NextToken == nil {
			break
		}

		req.NextToken = res.NextToken
	}

	sort.Slice(pss, func(i, j int) bool {
		if pi, pj := pss[i], pss[j]; pi.Service == pj.Service {
			return pi.Started.Before(pj.Started)
		} else {
			return pi.Service < pj.Service
		}
	})

	return pss, nil
}

func (p *Provider) processFromTask(app string, t *ecs.Task) (*types.Process, error) {
	ap := strings.Split(*t.TaskArn, "/")
	id := ap[len(ap)-1]

	td, err := p.fetchTaskDefinition(*t.TaskDefinitionArn)
	if err != nil {
		return nil, err
	}

	if len(td.ContainerDefinitions) < 1 {
		return nil, fmt.Errorf("no container for %s", *t.TaskDefinitionArn)
	}

	cd := *td.ContainerDefinitions[0]

	labels := map[string]string{}

	for k, v := range cd.DockerLabels {
		labels[k] = *v
	}

	cp := make([]string, len(cd.Command))

	for i, c := range cd.Command {
		cp[i] = *c
	}

	if len(cp) >= 2 && cp[0] == "sh" && cp[1] == "-c" {
		cp = cp[2:]
	}

	ps := &types.Process{
		Id:      id,
		App:     labels["convox.app"],
		Command: shellquote.Join(cp...),
		Release: labels["convox.release"],
		Service: labels["convox.service"],
		Status:  strings.ToLower(*t.LastStatus),
		Type:    labels["convox.type"],
	}

	if t.StartedAt != nil {
		ps.Started = *t.StartedAt
	}

	return ps, nil
}

func (p *Provider) fetchTaskDefinition(arn string) (*ecs.TaskDefinition, error) {
	res, err := p.ECS().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(arn),
	})
	if err != nil {
		return nil, err
	}

	return res.TaskDefinition, nil
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
