package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/template"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
	"github.com/fsouza/go-dockerclient"
)

const (
	printableTime = "2006-01-02 15:04:05"
	sortableTime  = "20060102.150405.000000000"
)

type Provider struct {
	Config      *aws.Config
	Development bool
	Name        string
	Password    string
	Region      string
	Session     *session.Session
	Version     string
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func FromEnv() (*Provider, error) {
	session, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	region := os.Getenv("AWS_REGION")

	p := &Provider{
		Config:      &aws.Config{Region: aws.String(region)},
		Development: os.Getenv("DEVELOPMENT") == "true",
		Name:        os.Getenv("NAME"),
		Password:    os.Getenv("PASSWORD"),
		Region:      region,
		Session:     session,
		Version:     os.Getenv("VERSION"),
	}

	if p.Version == "" {
		if v, err := p.rackOutput("Version"); err == nil && v != "" {
			p.Version = v
		}
	}

	return p, nil
}

func (p *Provider) CloudFormation() *cloudformation.CloudFormation {
	return cloudformation.New(p.Session, p.Config)
}

func (p *Provider) CloudWatch() *cloudwatch.CloudWatch {
	return cloudwatch.New(p.Session, p.Config)
}

func (p *Provider) CloudWatchLogs() *cloudwatchlogs.CloudWatchLogs {
	return cloudwatchlogs.New(p.Session, p.Config)
}

func (p *Provider) Docker(host string) (*docker.Client, error) {
	return docker.NewClient(host)
}

func (p *Provider) ECR() *ecr.ECR {
	return ecr.New(p.Session, p.Config)
}

func (p *Provider) ECS() *ecs.ECS {
	return ecs.New(p.Session, p.Config)
}

func (p *Provider) EC2() *ec2.EC2 {
	return ec2.New(p.Session, p.Config)
}

func (p *Provider) KMS() *kms.KMS {
	return kms.New(p.Session, p.Config)
}

func (p *Provider) IAM() *iam.IAM {
	return iam.New(p.Session, p.Config)
}

func (p *Provider) S3() *s3.S3 {
	return s3.New(p.Session, p.Config)
}

func (p *Provider) SimpleDB() *simpledb.SimpleDB {
	return simpledb.New(p.Session, p.Config)
}

func (p *Provider) SQS() *sqs.SQS {
	return sqs.New(p.Session, p.Config)
}

func (p *Provider) STS() *sts.STS {
	return sts.New(p.Session, p.Config)
}

func awsError(err error) string {
	if ae, ok := err.(awserr.Error); ok {
		return ae.Code()
	}
	return ""
}

func formationTemplate(name string, data interface{}) ([]byte, error) {
	var buf bytes.Buffer

	tn := fmt.Sprintf("%s.json.tmpl", name)
	tf := fmt.Sprintf("provider/aws/formation/%s", tn)

	t, err := template.New(tn).Funcs(formationHelpers()).ParseFiles(tf)
	if err != nil {
		return nil, err
	}

	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	var v interface{}

	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		switch t := err.(type) {
		case *json.SyntaxError:
			return nil, jsonSyntaxError(t, buf.Bytes())
		}
		return nil, err
	}

	return json.MarshalIndent(v, "", "  ")
}

func jsonSyntaxError(err *json.SyntaxError, data []byte) error {
	start := bytes.LastIndex(data[:err.Offset], []byte("\n")) + 1
	line := bytes.Count(data[:start], []byte("\n"))
	pos := int(err.Offset) - start - 1
	ltext := strings.Split(string(data), "\n")[line]

	return fmt.Errorf("json syntax error: line %d pos %d: %s: %s", line, pos, err.Error(), ltext)
}

type target struct {
	Balancer string
	Endpoint string
	Port     string
}

func formationHelpers() template.FuncMap {
	return template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"priority": func(app, service string) uint32 {
			return crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s-%s", app, service))) % 50000
		},
		"resource": func(s string) string {
			return upperName(s)
		},
		"target": func(m *manifest.Manifest, service string) (*target, error) {
			for _, b := range m.Balancers {
				for _, e := range b.Endpoints {
					if e.Target == "" {
						continue
					}

					u, err := url.Parse(e.Target)
					if err != nil {
						return nil, err
					}

					return &target{Balancer: b.Name, Endpoint: e.Port, Port: u.Port()}, nil
				}
			}

			return nil, fmt.Errorf("no target found")
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}
}

func (p *Provider) accountID() (string, error) {
	res, err := p.STS().GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	return *res.Account, nil
}

func (p *Provider) appOutput(app string, resource string) (string, error) {
	return p.stackOutput(fmt.Sprintf("%s-%s", p.Name, app), resource)
}

func (p *Provider) appResource(app string, resource string) (string, error) {
	return p.stackResource(fmt.Sprintf("%s-%s", p.Name, app), resource)
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

	group, err := p.appResource(app, "Logs")
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

func humanStatus(status string) string {
	switch status {
	case "CREATE_COMPLETE":
		return "running"
	case "CREATE_IN_PROGRESS":
		return "creating"
	case "DELETE_IN_PROGRESS":
		return "deleting"
	case "DELETE_FAILED":
		return "error"
	case "ROLLBACK_COMPLETE":
		return "running"
	case "ROLLBACK_IN_PROGRESS":
		return "rollback"
	case "UPDATE_COMPLETE":
		return "running"
	case "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS":
		return "updating"
	case "UPDATE_IN_PROGRESS":
		return "updating"
	case "UPDATE_ROLLBACK_COMPLETE":
		return "running"
	case "UPDATE_ROLLBACK_IN_PROGRESS":
		return "rollback"
	case "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS":
		return "rollback"
	default:
		return status
	}
}

func (p *Provider) rackOutput(output string) (string, error) {
	return p.stackOutput(p.Name, output)
}

func (p *Provider) rackResource(resource string) (string, error) {
	return p.stackResource(p.Name, resource)
}

func (p *Provider) stackOutput(name string, output string) (string, error) {
	res, err := p.CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if err != nil {
		return "", err
	}
	if len(res.Stacks) < 1 {
		return "", fmt.Errorf("no such stack: %s", name)
	}

	for _, o := range res.Stacks[0].Outputs {
		if *o.OutputKey == output {
			return *o.OutputValue, nil
		}
	}

	return "", fmt.Errorf("no such output for stack %s: %s", name, output)
}

func (p *Provider) stackResource(name string, resource string) (string, error) {
	res, err := p.CloudFormation().DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		LogicalResourceId: aws.String(resource),
		StackName:         aws.String(name),
	})
	if err != nil {
		return "", err
	}

	return *res.StackResourceDetail.PhysicalResourceId, nil
}

func (p *Provider) taskDefinition(app string, opts types.ProcessRunOptions) (string, error) {
	logs, err := p.appResource(app, "Logs")
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
				MemoryReservation: aws.Int64(256),
				Name:              aws.String(opts.Service),
			},
		},
		Family: aws.String(fmt.Sprintf("%s-%s", p.Name, app)),
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

	endpoint, err := p.stackOutput(p.Name, "Endpoint")
	if err != nil {
		return "", err
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword(p.Password, "")

	aenv := map[string]string{
		"APP":      app,
		"RACK_URL": u.String(),
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
		Family:  aws.String(fmt.Sprintf("%s-%s", p.Name, app)),
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

func upperName(name string) string {
	// myapp -> Myapp; my-app -> MyApp
	us := strings.ToUpper(name[0:1]) + name[1:]

	for {
		i := strings.Index(us, "-")

		if i == -1 {
			break
		}

		s := us[0:i]

		if len(us) > i+1 {
			s += strings.ToUpper(us[i+1 : i+2])
		}

		if len(us) > i+2 {
			s += us[i+2:]
		}

		us = s
	}

	return us
}
