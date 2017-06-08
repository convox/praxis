package aws

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/convox/praxis/types"
	docker "github.com/fsouza/go-dockerclient"
	shellquote "github.com/kballard/go-shellquote"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p *Provider) ProcessExec(app, pid, command string, opts types.ProcessExecOptions) (int, error) {
	t, err := p.taskForPid(pid)
	if err != nil {
		return 0, err
	}

	ci, err := p.containerInstance(*t.ContainerInstanceArn)
	if err != nil {
		return 0, err
	}

	ei, err := p.ec2Instance(*ci.Ec2InstanceId)
	if err != nil {
		return 0, err
	}

	host := fmt.Sprintf("http://%s:2376", *ei.PrivateIpAddress)

	if p.Development {
		host = fmt.Sprintf("http://%s:2376", *ei.PublicIpAddress)
	}

	dc, err := docker.NewClient(host)
	if err != nil {
		return 0, err
	}

	cs, err := dc.ListContainers(docker.ListContainersOptions{
		Filters: map[string][]string{
			"label": {fmt.Sprintf("com.amazonaws.ecs.task-arn=%s", *t.TaskArn)},
		},
	})
	if err != nil {
		return 0, err
	}
	if len(cs) < 1 {
		return 0, fmt.Errorf("could not find container for pid: %s", pid)
	}

	eres, err := dc.CreateExec(docker.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"sh", "-c", command},
		Container:    cs[0].ID,
	})
	if err != nil {
		return 0, err
	}

	success := make(chan struct{})

	go func() {
		<-success
		dc.ResizeExecTTY(eres.ID, opts.Height, opts.Width)
		success <- struct{}{}
	}()

	err = dc.StartExec(eres.ID, docker.StartExecOptions{
		Detach:       false,
		Tty:          true,
		RawTerminal:  true,
		InputStream:  ioutil.NopCloser(opts.Input),
		OutputStream: opts.Output,
		ErrorStream:  opts.Output,
		Success:      success,
	})
	if err != nil {
		return 0, err
	}

	ires, err := dc.InspectExec(eres.ID)
	if err != nil {
		return 0, err
	}

	return ires.ExitCode, nil
}

func (p *Provider) ProcessGet(app, pid string) (*types.Process, error) {
	t, err := p.taskForPid(pid)
	if err != nil {
		return nil, err
	}

	ps, err := p.processFromTask(app, t)
	if err != nil {
		return nil, err
	}

	if ps.App != app {
		return nil, fmt.Errorf("process not found: %s\n", pid)
	}

	return ps, nil
}

func (p *Provider) ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error) {
	tasks, err := p.rackTasks()
	if err != nil {
		return nil, err
	}

	pss := types.Processes{}

	for _, t := range tasks {
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

	sort.Slice(pss, func(i, j int) bool {
		if pi, pj := pss[i], pss[j]; pi.Service == pj.Service {
			return pi.Started.Before(pj.Started)
		} else {
			return pi.Service < pj.Service
		}
	})

	return pss, nil
}

func (p *Provider) ProcessLogs(app, pid string, opts types.LogsOptions) (io.ReadCloser, error) {
	group, err := p.appResource(app, "Logs")
	if err != nil {
		return nil, err
	}

	t, err := p.taskForPid(pid)
	if err != nil {
		return nil, err
	}
	if len(t.Containers) != 1 {
		return nil, fmt.Errorf("invalid container for task: %s\n", pid)
	}

	stream := fmt.Sprintf("convox/%s/%s", *t.Containers[0].Name, pid)

	r, w := io.Pipe()

	go p.subscribeLogs(group, stream, opts, w)

	go func() {
		for {
			t, err := p.taskForPid(pid)
			if err != nil {
				w.Close()
				return
			}

			if *t.LastStatus == "STOPPED" {
				w.Close()
				return
			}

			time.Sleep(2 * time.Second)
		}
	}()

	return r, nil
}

func (p *Provider) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return 0, err
	}

	pid, err := p.ProcessStart(app, opts)
	if err != nil {
		return 0, err
	}

	treq := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []*string{aws.String(pid)},
	}

	if opts.Output != nil {
		if err := p.ECS().WaitUntilTasksRunning(treq); err != nil {
			return 0, err
		}

		return p.ProcessExec(app, pid, opts.Command, types.ProcessExecOptions{
			Height: opts.Height,
			Width:  opts.Width,
			Input:  opts.Input,
			Output: opts.Output,
		})
	}

	if err := p.ECS().WaitUntilTasksStopped(treq); err != nil {
		return 0, err
	}

	tres, err := p.ECS().DescribeTasks(treq)
	if err != nil {
		return 0, err
	}
	if len(tres.Tasks) != 1 {
		return 0, fmt.Errorf("unable to find process")
	}
	if len(tres.Tasks[0].Containers) != 1 {
		return 0, fmt.Errorf("no container for pid: %s", pid)
	}

	return int(*tres.Tasks[0].Containers[0].ExitCode), nil
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
		TaskDefinition: aws.String(td),
	}

	res, err := p.ECS().RunTask(req)
	if err != nil {
		return "", err
	}
	if len(res.Tasks) != 1 {
		if len(res.Failures) > 0 {
			return "", fmt.Errorf("unable to start process: %s", *res.Failures[0].Reason)
		}
		return "", fmt.Errorf("unable to start process")
	}

	parts := strings.Split(*res.Tasks[0].TaskArn, "/")
	pid := parts[len(parts)-1]

	return pid, nil
}

func (p *Provider) ProcessStop(app, pid string) error {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return err
	}

	_, err = p.ECS().StopTask(&ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Reason:  aws.String("ProcessStop"),
		Task:    aws.String(pid),
	})
	if err != nil {
		return err
	}

	return nil
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

func (p *Provider) rackTasks() ([]*ecs.Task, error) {
	tasks := []*ecs.Task{}

	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

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

		tasks = append(tasks, dres.Tasks...)

		if res.NextToken == nil {
			break
		}

		req.NextToken = res.NextToken
	}

	return tasks, nil
}
