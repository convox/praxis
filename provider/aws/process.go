package aws

import (
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/convox/praxis/types"
	docker "github.com/fsouza/go-dockerclient"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (p *Provider) ProcessGet(app, pid string) (*types.Process, error) {
	return nil, nil
}

func (p *Provider) ProcessList(app string, opts types.ProcessListOptions) (types.Processes, error) {
	return nil, nil
}

func (p *Provider) ProcessLogs(app, pid string) (io.ReadCloser, error) {
	r, w := io.Pipe()

	go p.cloudwatchLogStream(app, pid, w)

	return r, nil
}

func (p *Provider) cloudwatchLogStream(app, pid string, w io.WriteCloser) {
	defer w.Close()

	task, err := p.taskForPid(app, pid)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return
	}

	arn := *task.TaskArn

	if len(task.Containers) < 1 {
		w.Write([]byte(fmt.Sprintf("no container for task: %s", arn)))
		return
	}

	ap := strings.Split(arn, "/")
	uuid := ap[len(ap)-1]
	name := *task.Containers[0].Name
	stream := fmt.Sprintf("convox/%s/%s", name, uuid)

	group, err := p.appResource(app, "LogGroup")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: %s\n", err)))
		return
	}

	req := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		StartFromHead: aws.Bool(true),
	}

	empty := 0

	for {
		res, err := p.CloudWatchLogs().GetLogEvents(req)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error: %s\n", err)))
			return
		}

		if len(res.Events) == 0 {
			if _, err := p.taskForPid(app, pid); err != nil {
				empty++
				if empty >= 4 {
					return
				}
			}
		} else {
			empty = 0
		}

		events := res.Events

		sort.Slice(events, func(i, j int) bool { return *events[i].Timestamp < *events[j].Timestamp })

		for _, e := range events {
			w.Write([]byte(fmt.Sprintf("%s\n", *e.Message)))
		}

		req.NextToken = res.NextForwardToken

		time.Sleep(250 * time.Millisecond)
	}
}

func (p *Provider) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	return 0, nil
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
	return nil
}

func (p *Provider) containerForPid(app, pid string) (*docker.Client, *docker.APIContainers, error) {
	task, err := p.taskForPid(app, pid)
	if err != nil {
		return nil, nil, err
	}

	host, err := p.dockerHostForInstance(*task.ContainerInstanceArn)
	if err != nil {
		return nil, nil, err
	}

	dc, err := p.Docker(host)
	if err != nil {
		return nil, nil, err
	}

	cs, err := dc.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"label": {fmt.Sprintf("com.amazonaws.ecs.task-arn=%s", *task.TaskArn)},
		},
	})
	if err != nil {
		return nil, nil, err
	}
	if len(cs) < 1 {
		return nil, nil, fmt.Errorf("no container for task: %s", *task.TaskArn)
	}

	return dc, &cs[0], nil
}

func (p *Provider) dockerHostForInstance(instance string) (string, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return "", err
	}

	req := &ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(cluster),
		ContainerInstances: []*string{aws.String(instance)},
	}

	res, err := p.ECS().DescribeContainerInstances(req)
	if err != nil {
		return "", err
	}

	if len(res.ContainerInstances) < 1 {
		return "", fmt.Errorf("no such container instance: %s", instance)
	}

	ereq := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(*res.ContainerInstances[0].Ec2InstanceId)},
	}

	eres, err := p.EC2().DescribeInstances(ereq)
	if err != nil {
		return "", err
	}

	if len(eres.Reservations) != 1 || len(eres.Reservations[0].Instances) != 1 {
		return "", fmt.Errorf("could not find instance: %s", *ereq.InstanceIds[0])
	}

	i := eres.Reservations[0].Instances[0]

	host := fmt.Sprintf("http://%s:2376", *i.PrivateIpAddress)

	if p.Development {
		host = fmt.Sprintf("http://%s:2376", *i.PublicIpAddress)
	}

	return host, nil
}

func (p *Provider) taskDefinition(app string, opts types.ProcessRunOptions) (string, error) {
	logs, err := p.appResource(app, "LogGroup")
	if err != nil {
		return "", err
	}

	req := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Cpu:       aws.Int64(512),
				Essential: aws.Bool(true),
				Image:     aws.String(""),
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String("awslogs"),
					Options: map[string]*string{
						"awslogs-region":        aws.String(p.Region),
						"awslogs-group":         aws.String(logs),
						"awslogs-stream-prefix": aws.String("convox"),
					},
				},
				MemoryReservation: aws.Int64(512),
				Name:              aws.String(opts.Service),
			},
		},
		Family: aws.String(fmt.Sprintf("%s-%s", p.Rack, app)),
	}

	if opts.Command != "" {
		req.ContainerDefinitions[0].Command = []*string{aws.String("sh"), aws.String("-c"), aws.String(opts.Command)}
	}

	for k, v := range opts.Environment {
		req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	aenv := map[string]string{
		"APP":      app,
		"RACK_URL": "https://david-praxis.ngrok.io",
	}

	for k, v := range aenv {
		req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	if opts.Image != "" {
		req.ContainerDefinitions[0].Image = aws.String(opts.Image)
	}

	for from, to := range opts.Ports {
		req.ContainerDefinitions[0].PortMappings = append(req.ContainerDefinitions[0].PortMappings, &ecs.PortMapping{
			HostPort:      aws.Int64(int64(from)),
			ContainerPort: aws.Int64(int64(to)),
		})
	}

	i := 0

	for from, to := range opts.Volumes {
		name := fmt.Sprintf("volume-%d", i)

		req.ContainerDefinitions[0].MountPoints = append(req.ContainerDefinitions[0].MountPoints, &ecs.MountPoint{
			ContainerPath: aws.String(to),
			SourceVolume:  aws.String(name),
		})

		req.Volumes = append(req.Volumes, &ecs.Volume{
			Host: &ecs.HostVolumeProperties{
				SourcePath: aws.String(from),
			},
			Name: aws.String(name),
		})
	}

	res, err := p.ECS().RegisterTaskDefinition(req)
	if err != nil {
		return "", err
	}

	return *res.TaskDefinition.TaskDefinitionArn, nil
}

func (p *Provider) taskForPid(app, pid string) (*ecs.Task, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

	req := &ecs.ListTasksInput{
		Cluster: aws.String(cluster),
		Family:  aws.String(fmt.Sprintf("%s-%s", p.Rack, app)),
	}

	var task *ecs.Task

	err = p.ECS().ListTasksPages(req, func(res *ecs.ListTasksOutput, last bool) bool {
		for _, arn := range res.TaskArns {
			parts := strings.Split(*arn, "-")
			if parts[len(parts)-1] != pid {
				continue
			}

			req := &ecs.DescribeTasksInput{
				Cluster: aws.String(cluster),
				Tasks:   []*string{arn},
			}

			res, err := p.ECS().DescribeTasks(req)
			if err != nil {
				return false
			}

			if len(res.Tasks) == 1 {
				task = res.Tasks[0]
				return false
			}
		}

		return true
	})
	if err != nil {
		return nil, err
	}

	if task != nil {
		return task, nil
	}

	return nil, fmt.Errorf("could not find task for pid: %s", pid)
}
