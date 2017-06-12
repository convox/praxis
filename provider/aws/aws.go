package aws

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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
	"github.com/convox/praxis/helpers"
	"github.com/convox/praxis/manifest"
	"github.com/convox/praxis/types"
	"github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

const ()

type Provider struct {
	Config      *aws.Config
	Context     context.Context
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
		Context:     context.Background(),
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

func (p *Provider) AutoScaling() *autoscaling.AutoScaling {
	return autoscaling.New(p.Session, p.Config)
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

func (p *Provider) clusterServices() ([]*ecs.Service, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

	req := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	}

	ss := []*ecs.Service{}

	for {
		res, err := p.ECS().ListServices(req)
		if err != nil {
			return nil, err
		}

		sres, err := p.ECS().DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  aws.String(cluster),
			Services: res.ServiceArns,
		})
		if err != nil {
			return nil, err
		}

		ss = append(ss, sres.Services...)

		if res.NextToken == nil {
			break
		}

		req.NextToken = res.NextToken
	}

	return ss, nil
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

func (p *Provider) containerInstances() ([]*ecs.ContainerInstance, error) {
	cluster, err := p.rackResource("RackCluster")
	if err != nil {
		return nil, err
	}

	ii := []*ecs.ContainerInstance{}

	req := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(cluster),
	}

	for {
		res, err := p.ECS().ListContainerInstances(req)
		if err != nil {
			return nil, err
		}

		ires, err := p.ECS().DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(cluster),
			ContainerInstances: res.ContainerInstanceArns,
		})
		if err != nil {
			return nil, err
		}

		ii = append(ii, ires.ContainerInstances...)

		if res.NextToken == nil {
			break
		}

		req.NextToken = res.NextToken
	}

	return ii, nil
}

func (p *Provider) deleteBucket(bucket string) error {
	req := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}

	err := p.S3().ListObjectsPages(req, func(res *s3.ListObjectsOutput, last bool) bool {
		objects := make([]*s3.ObjectIdentifier, len(res.Contents))

		for i, o := range res.Contents {
			objects[i] = &s3.ObjectIdentifier{Key: o.Key}
		}

		if len(objects) == 0 {
			return false
		}

		p.S3().DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &s3.Delete{Objects: objects},
		})

		return true
	})
	if err != nil {
		return err
	}

	_, err = p.S3().DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) describeStack(name string) (*cloudformation.Stack, error) {
	if v, ok := cache.Get("describeStack", name).(*cloudformation.Stack); ok {
		return v, nil
	}

	res, err := p.CloudFormation().DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}
	if len(res.Stacks) < 1 {
		return nil, fmt.Errorf("no such stack: %s", name)
	}

	if err := cache.Set("describeStack", name, res.Stacks[0], 10*time.Second); err != nil {
		return nil, err
	}

	return res.Stacks[0], nil
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

func (p *Provider) writeLogf(group, stream, format string, args ...interface{}) error {
	req := &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			&cloudwatchlogs.InputLogEvent{
				Message:   aws.String(fmt.Sprintf(format, args...)),
				Timestamp: aws.Int64(time.Now().UTC().UnixNano() / 1000000),
			},
		},
	}

	for {
		res, err := p.CloudWatchLogs().PutLogEvents(req)
		switch awsError(err) {
		// if the log stream doesnt exist, create it
		case "ResourceNotFoundException":
			_, err := p.CloudWatchLogs().CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
				LogGroupName:  aws.String(group),
				LogStreamName: aws.String(stream),
			})
			if err != nil {
				return err
			}
			continue
		// need to set the sequence token
		case "DataAlreadyAcceptedException":
			req.SequenceToken = res.NextSequenceToken
			continue
		case "InvalidSequenceTokenException":
			if ae, ok := err.(awserr.Error); ok {
				parts := strings.Split(ae.Message(), " ")
				req.SequenceToken = aws.String(parts[len(parts)-1])
				continue
			}
		}

		return err
	}
}

func (p *Provider) subscribeLogs(group, stream string, opts types.LogsOptions, w io.WriteCloser) error {
	return p.subscribeLogsCallback(group, stream, opts, w, nil)
}

func (p *Provider) subscribeLogsCallback(group, stream string, opts types.LogsOptions, w io.WriteCloser, fn func() bool) error {
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
		// Always make sure there is something we can write to
		if _, err := fmt.Fprintf(w, ""); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		events := []*cloudwatchlogs.FilteredLogEvent{}

		err := p.CloudWatchLogs().FilterLogEventsPages(req, func(res *cloudwatchlogs.FilterLogEventsOutput, last bool) bool {
			for _, e := range res.Events {
				events = append(events, e)
			}

			return true
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			break
		}

		sort.Slice(events, func(i, j int) bool { return *events[i].Timestamp < *events[j].Timestamp })

		for _, e := range events {
			parts := strings.SplitN(*e.LogStreamName, "/", 3)

			if len(parts) == 3 {
				pp := strings.Split(parts[2], "-")
				ts := time.Unix(*e.Timestamp/1000, *e.Timestamp%1000*1000).UTC()

				var err error

				if opts.Prefix {
					_, err = fmt.Fprintf(w, "%s %s/%s/%s %s\n", ts.Format(helpers.PrintableTime), parts[0], parts[1], pp[len(pp)-1], *e.Message)
				} else {
					_, err = fmt.Fprintf(w, "%s\n", *e.Message)
				}

				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
			}
		}

		if !opts.Follow {
			break
		}

		if fn != nil {
			if !fn() {
				return nil
			}
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

	stack, err := p.describeStack(name)
	if err != nil {
		return "", err
	}

	for _, o := range stack.Outputs {
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

func (p *Provider) fetchTaskDefinition(arn string) (*ecs.TaskDefinition, error) {
	if v, ok := cache.Get("taskDefinition", arn).(*ecs.TaskDefinition); ok {
		return v, nil
	}

	res, err := p.ECS().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(arn),
	})
	if err != nil {
		return nil, err
	}

	if err := cache.Set("taskDefinition", arn, res.TaskDefinition, 24*time.Hour); err != nil {
		return nil, err
	}

	return res.TaskDefinition, nil
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

	// get release and manifest for initial environment and volumes
	var m *manifest.Manifest
	var release *types.Release
	var service *manifest.Service

	if opts.Release != "" {
		m, release, err = helpers.ReleaseManifest(p, app, opts.Release)
		if err != nil {
			return "", errors.WithStack(err)
		}

		// if service is not defined in manifest, i.e. "build", carry on
		service, err = m.Service(opts.Service)
		if err != nil && !strings.Contains(err.Error(), "no such service") {
			return "", errors.WithStack(err)
		}
	}

	if service != nil {
		env, err := m.ServiceEnvironment(opts.Service)
		if err != nil {
			return "", err
		}

		for k, v := range env {
			req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
				Name:  aws.String(k),
				Value: aws.String(v),
			})
		}

		// resource environment
		rs, err := p.ResourceList(app)
		if err != nil {
			return "", err
		}

		for _, r := range rs {
			req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
				Name:  aws.String(strings.ToUpper(fmt.Sprintf("%s_URL", r.Name))),
				Value: aws.String(r.Endpoint),
			})
		}

		// volumes for service
		s, err := m.Service(opts.Service)
		if err != nil {
			return "", err
		}

		for i, v := range s.Volumes {
			var from, to string
			parts := strings.SplitN(v, ":", 2)
			switch len(parts) {
			case 1:
				from = path.Join("/volumes", v)
				to = v
			case 2:
				from = parts[0]
				to = parts[1]
			default:
				return "", fmt.Errorf("invalid volume definition: %s", v)

			}

			name := fmt.Sprintf("volume-%d", i) // manifest volumes

			req.Volumes = append(req.Volumes, &ecs.Volume{
				Host: &ecs.HostVolumeProperties{
					SourcePath: aws.String(from),
				},
				Name: aws.String(name),
			})

			req.ContainerDefinitions[0].MountPoints = append(req.ContainerDefinitions[0].MountPoints, &ecs.MountPoint{
				ContainerPath: aws.String(to),
				SourceVolume:  aws.String(name),
			})
		}
	}

	for k, v := range opts.Environment {
		req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	if opts.Command != "" {
		req.ContainerDefinitions[0].Command = []*string{aws.String("sh"), aws.String("-c"), aws.String(opts.Command)}
	}

	if opts.Output != nil {
		req.ContainerDefinitions[0].Command = []*string{aws.String("sleep"), aws.String("3600")}
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

	if opts.Service != "" && opts.Image == "" {
		if release == nil {
			return "", fmt.Errorf("no release for app: %s", app)
		}

		req.ContainerDefinitions[0].Environment = append(req.ContainerDefinitions[0].Environment, &ecs.KeyValuePair{
			Name:  aws.String("RELEASE"),
			Value: aws.String(release.Id),
		})

		account, err := p.accountID()
		if err != nil {
			return "", err
		}

		repo, err := p.appResource(app, "Repository")
		if err != nil {
			return "", err
		}

		req.ContainerDefinitions[0].Image = aws.String(fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s.%s", account, p.Region, repo, opts.Service, release.Build))
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
		name := fmt.Sprintf("volume-o-%d", i) // one-off volumes

		req.Volumes = append(req.Volumes, &ecs.Volume{
			Host: &ecs.HostVolumeProperties{
				SourcePath: aws.String(from),
			},
			Name: aws.String(name),
		})

		req.ContainerDefinitions[0].MountPoints = append(req.ContainerDefinitions[0].MountPoints, &ecs.MountPoint{
			ContainerPath: aws.String(to),
			SourceVolume:  aws.String(name),
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
