package aws

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/convox/praxis/types"
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
	task, err := p.taskForPid(app, pid)
	if err != nil {
		return nil, err
	}

	host, err := p.dockerHostForTask(*task.ContainerInstanceArn)
	if err != nil {
		return nil, err
	}

	fmt.Printf("host = %+v\n", host)

	return nil, nil
}

func (p *Provider) dockerHostForTask(instance string) (string, error) {
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

	fmt.Printf("ereq = %+v\n", ereq)
	// fmt.Printf("res = %+v\n", res)

	return "", nil
}

func (p *Provider) ProcessRun(app string, opts types.ProcessRunOptions) (int, error) {
	return 0, nil
}

func (p *Provider) ProcessStart(app string, opts types.ProcessStartOptions) (string, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return "", err
	}

	td, err := p.taskDefinition(app, opts.Command, opts.Environment, opts.Image, opts.Service, opts.Volumes)
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

// func (p *Provider) containerInstances(cluster string) ([]*string, error) {
//   req := &ecs.ListContainerInstancesInput{
//     Cluster: aws.String(cluster),
//   }

//   ci := []*string{}

//   err := p.ECS().ListContainerInstancesPages(req, func(res *ecs.ListContainerInstancesOutput, last bool) bool {
//     ci = append(ci, res.ContainerInstanceArns...)
//     return true
//   })
//   if err != nil {
//     return nil, err
//   }

//   rci := make([]*string, len(ci))

//   for i, v := range rand.Perm(len(ci)) {
//     rci[v] = ci[i]
//   }

//   return rci, nil
// }

func (p *Provider) taskDefinition(app string, command string, env map[string]string, image string, service string, volumes map[string]string) (string, error) {
	logs, err := p.appResource(app, "LogGroup")
	if err != nil {
		return "", err
	}

	req := &ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
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
				MemoryReservation: aws.Int64(128),
				Name:              aws.String(service),
			},
		},
		Family: aws.String(fmt.Sprintf("%s-%s", p.Rack, app)),
	}

	if command != "" {
		req.ContainerDefinitions[0].Command = []*string{aws.String("sh"), aws.String("-c"), aws.String(command)}
	}

	for k, v := range env {
		req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
		Name:  aws.String("RACK_URL"),
		Value: aws.String("https://david-praxis.ngrok.io"),
	})

	if image != "" {
		req.ContainerDefinitions[0].Image = aws.String(image)
	}

	i := 0

	for from, to := range volumes {
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
