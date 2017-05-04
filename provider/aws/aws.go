package aws

import (
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

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
	"github.com/convox/praxis/cache"
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

func (p *Provider) accountID() (string, error) {
	if v, ok := cache.Get("accountID", "").(string); ok {
		return v, nil
	}

	res, err := p.STS().GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}

	a := *res.Account

	if err := cache.Set("accountID", "", a, 24*time.Hour); err != nil {
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

func (p *Provider) containerInstance(arn string) (*ecs.ContainerInstance, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

	res, err := p.ECS().DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(cluster),
		ContainerInstances: []*string{aws.String(arn)},
	})
	if err != nil {
		return nil, err
	}
	if len(res.ContainerInstances) < 1 {
		return nil, fmt.Errorf("could not find container instance: %s", arn)
	}

	return res.ContainerInstances[0], nil
}

func (p *Provider) ec2Instance(id string) (*ec2.Instance, error) {
	res, err := p.EC2().DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(id)},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Reservations) < 1 || len(res.Reservations[0].Instances) < 1 {
		return nil, fmt.Errorf("could not find ec2 instance: %s", id)
	}

	return res.Reservations[0].Instances[0], nil
}

func (p *Provider) rackOutput(output string) (string, error) {
	return p.stackOutput(p.Name, output)
}

func (p *Provider) rackResource(resource string) (string, error) {
	return p.stackResource(p.Name, resource)
}

func (p *Provider) subscribeLogs(group, stream string, opts types.LogsOptions, w io.WriteCloser) error {
	defer w.Close()

	req := &cloudwatchlogs.FilterLogEventsInput{
		Interleaved:  aws.Bool(true),
		LogGroupName: aws.String(group),
	}

	if stream != "" {
		req.LogStreamNames = []*string{aws.String(stream)}
	}

	if opts.Filter != "" {
		req.FilterPattern = aws.String(opts.Filter)
	}

	if !opts.Since.IsZero() {
		req.StartTime = aws.Int64(opts.Since.UTC().Unix() * 1000)
	}

	for {
		events := []*cloudwatchlogs.FilteredLogEvent{}

		err := p.CloudWatchLogs().FilterLogEventsPages(req, func(res *cloudwatchlogs.FilterLogEventsOutput, last bool) bool {
			for _, e := range res.Events {
				events = append(events, e)
			}

			return true
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}

		sort.Slice(events, func(i, j int) bool { return *events[i].Timestamp < *events[j].Timestamp })

		for _, e := range events {
			parts := strings.SplitN(*e.LogStreamName, "/", 3)

			if len(parts) == 3 {
				pp := strings.Split(parts[2], "-")
				ts := time.Unix(*e.Timestamp/1000, *e.Timestamp%1000*1000).UTC()

				fmt.Fprintf(w, "%s %s/%s/%s %s\n", ts.Format(printableTime), parts[0], parts[1], pp[len(pp)-1], *e.Message)
			}
		}

		if !opts.Follow {
			break
		}

		time.Sleep(1 * time.Second)

		if len(events) > 0 {
			req.StartTime = aws.Int64(*events[len(events)-1].Timestamp + 1)
		}
	}

	return nil
}

func (p *Provider) stackOutput(name string, output string) (string, error) {
	ck := fmt.Sprintf("%s/%s", name, output)

	if v, ok := cache.Get("stackOutput", ck).(string); ok {
		return v, nil
	}

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
			ov := *o.OutputValue

			if err := cache.Set("stackOutput", ck, ov, 1*time.Minute); err != nil {
				return "", err
			}

			return ov, nil
		}
	}

	return "", fmt.Errorf("no such output for stack %s: %s", name, output)
}

func (p *Provider) stackResource(name string, resource string) (string, error) {
	ck := fmt.Sprintf("%s/%s", name, resource)

	if v, ok := cache.Get("stackResource", ck).(string); ok {
		return v, nil
	}

	res, err := p.CloudFormation().DescribeStackResource(&cloudformation.DescribeStackResourceInput{
		LogicalResourceId: aws.String(resource),
		StackName:         aws.String(name),
	})
	if err != nil {
		return "", err
	}

	r := *res.StackResourceDetail.PhysicalResourceId

	if err := cache.Set("stackResource", ck, r, 1*time.Minute); err != nil {
		return "", err
	}

	return r, nil
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

	if opts.Stream != nil {
		req.ContainerDefinitions[0].Command = []*string{aws.String("sleep"), aws.String("3600")}
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

	if opts.Service != "" {
		account, err := p.rackOutput("Account")
		if err != nil {
			return "", err
		}

		repo, err := p.appResource(app, "Repository")
		if err != nil {
			return "", err
		}

		rs, err := p.ReleaseList(app, types.ReleaseListOptions{Count: 1})
		if err != nil {
			return "", err
		}

		if len(rs) < 1 {
			r, err := p.releaseFork(app)
			if err != nil {
				return "", err
			}

			rs = append(rs, *r)
		}

		req.ContainerDefinitions[0].Image = aws.String(fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s.%s", account, p.Region, repo, opts.Service, rs[0].Build))
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

func (p *Provider) taskForPid(pid string) (*ecs.Task, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

	res, err := p.ECS().DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []*string{aws.String(pid)},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Tasks) < 1 {
		return nil, fmt.Errorf("could not find task for pid: %s", pid)
	}

	return res.Tasks[0], nil
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
