package main

import (
	"fmt"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	awsdynamodb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	awsec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	awsecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	awsecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	awslogs "github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	awssns "github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	awssqs "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ServiceStackProps struct {
	awscdk.StackProps

	ServiceName   string
	EcrRepoName   string
	ImageTag      string
	ContainerPort int    // default 8888
	HealthPath    string // default /health
	DesiredCount  int    // default 1
	InternalOnly  bool   // default true

	// DDB table names to import and grant to the task
	CountersTableName string
	RomancesTableName string

	// Optional messaging, if your app publishes/consumes
	TopicArn string // to allow sns:Publish
	QueueArn string // to allow sqs:Receive/Delete
	QueueUrl string // passed to container for consumers
}

func NewServiceStack(scope constructs.Construct, id string, props *ServiceStackProps) awscdk.Stack {
	var sp awscdk.StackProps
	if props != nil {
		sp = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sp)

	// defaults
	if props.ContainerPort == 0 {
		props.ContainerPort = 8888
	}
	if props.HealthPath == "" {
		props.HealthPath = "/health"
	}
	if props.DesiredCount == 0 {
		props.DesiredCount = 1
	}
	// don't force InternalOnly; honor the prop

	// VPC
	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{MaxAzs: jsii.Number(2), NatGateways: jsii.Number(1)})

	// SGs
	albSg := awsec2.NewSecurityGroup(stack, jsii.String("AlbSg"), &awsec2.SecurityGroupProps{Vpc: vpc, AllowAllOutbound: jsii.Bool(true)})
	svcSg := awsec2.NewSecurityGroup(stack, jsii.String("SvcSg"), &awsec2.SecurityGroupProps{Vpc: vpc, AllowAllOutbound: jsii.Bool(true)})
	awsec2.NewCfnSecurityGroupIngress(stack, jsii.String("AlbToSvcPort"), &awsec2.CfnSecurityGroupIngressProps{
		GroupId: svcSg.SecurityGroupId(), SourceSecurityGroupId: albSg.SecurityGroupId(),
		IpProtocol: jsii.String("tcp"), FromPort: jsii.Number(props.ContainerPort), ToPort: jsii.Number(props.ContainerPort),
	})

	// ECS cluster + task role
	cluster := awsecs.NewCluster(stack, jsii.String("Cluster"), &awsecs.ClusterProps{Vpc: vpc})
	taskRole := awsiam.NewRole(stack, jsii.String("TaskRole"), &awsiam.RoleProps{AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil)})

	// Import DDB tables and grant to the role
	if props.CountersTableName != "" {
		t := awsdynamodb.Table_FromTableName(stack, jsii.String("CountersImport"), jsii.String(props.CountersTableName))
		t.GrantReadWriteData(taskRole)
	}
	if props.RomancesTableName != "" {
		t := awsdynamodb.Table_FromTableName(stack, jsii.String("RomancesImport"), jsii.String(props.RomancesTableName))
		t.GrantReadWriteData(taskRole)
	}

	// Optional: messaging permissions
	if props.TopicArn != "" {
		topic := awssns.Topic_FromTopicArn(stack, jsii.String("AppTopicImported"), jsii.String(props.TopicArn))
		// Least-privilege: allow only Publish to this topic
		topic.GrantPublish(taskRole)
	}
	if props.QueueArn != "" {
		queue := awssqs.Queue_FromQueueAttributes(stack, jsii.String("AppQueueImported"), &awssqs.QueueAttributes{QueueArn: jsii.String(props.QueueArn)})
		// Permissions for polling workers: Receive + Delete + GetAttributes
		queue.GrantConsumeMessages(taskRole)
	}

	// Logs
	logGroup := awslogs.NewLogGroup(stack, jsii.String("AppLogs"), &awslogs.LogGroupProps{Retention: awslogs.RetentionDays_ONE_WEEK})

	// Image
	repo := awsecr.Repository_FromRepositoryName(stack, jsii.String("Repo"), jsii.String(props.EcrRepoName))
	image := awsecs.ContainerImage_FromEcrRepository(repo, jsii.String(props.ImageTag))

	// Task def + container
	td := awsecs.NewFargateTaskDefinition(stack, jsii.String("TaskDef"), &awsecs.FargateTaskDefinitionProps{Cpu: jsii.Number(256), MemoryLimitMiB: jsii.Number(512), TaskRole: taskRole})
	container := td.AddContainer(jsii.String("App"), &awsecs.ContainerDefinitionOptions{
		Image:        image,
		PortMappings: &[]*awsecs.PortMapping{{ContainerPort: jsii.Number(props.ContainerPort)}},
		Environment: &map[string]*string{
			"AWS_REGION":   awscdk.Stack_Of(stack).Region(),
			"SERVICE_NAME": jsii.String(props.ServiceName),
			"PORT":         jsii.String(fmt.Sprintf("%d", props.ContainerPort)),
			"DDB_COUNTERS": jsii.String(props.CountersTableName),
			"DDB_ROMANCES": jsii.String(props.RomancesTableName),
			// Messaging hints for the app if it needs them
			"SNS_TOPIC_ARN": jsii.String(props.TopicArn),
			"SQS_QUEUE_ARN": jsii.String(props.QueueArn),
			"SQS_QUEUE_URL": jsii.String(props.QueueUrl),
		},
		Logging: awsecs.LogDriver_AwsLogs(&awsecs.AwsLogDriverProps{LogGroup: logGroup, StreamPrefix: jsii.String("app")}),
	})
	_ = container

	// Service
	svc := awsecs.NewFargateService(stack, jsii.String("Service"), &awsecs.FargateServiceProps{
		Cluster:        cluster,
		ServiceName:    jsii.String(props.ServiceName),
		TaskDefinition: td,
		DesiredCount:   jsii.Number(props.DesiredCount),
		AssignPublicIp: jsii.Bool(!props.InternalOnly && true),
		SecurityGroups: &[]awsec2.ISecurityGroup{svcSg},
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: func() awsec2.SubnetType {
			if props.InternalOnly {
				return awsec2.SubnetType_PRIVATE_WITH_EGRESS
			}
			return awsec2.SubnetType_PUBLIC
		}()},
		CircuitBreaker:         &awsecs.DeploymentCircuitBreaker{Rollback: jsii.Bool(true)},
		HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(60)),
	})

	// ALB
	alb := elbv2.NewApplicationLoadBalancer(stack, jsii.String("Alb"), &elbv2.ApplicationLoadBalancerProps{
		Vpc: vpc, InternetFacing: jsii.Bool(!props.InternalOnly), SecurityGroup: albSg,
		LoadBalancerName: jsii.String(props.ServiceName + "-alb"),
		VpcSubnets: &awsec2.SubnetSelection{SubnetType: func() awsec2.SubnetType {
			if props.InternalOnly {
				return awsec2.SubnetType_PRIVATE_WITH_EGRESS
			}
			return awsec2.SubnetType_PUBLIC
		}()},
	})
	ln := alb.AddListener(jsii.String("Http"), &elbv2.BaseApplicationListenerProps{Port: jsii.Number(80), Open: jsii.Bool(!props.InternalOnly)})

	tg := elbv2.NewApplicationTargetGroup(stack, jsii.String("AppTg"), &elbv2.ApplicationTargetGroupProps{
		Vpc:         vpc,
		Port:        jsii.Number(props.ContainerPort),
		Protocol:    elbv2.ApplicationProtocol_HTTP,
		TargetType:  elbv2.TargetType_IP,
		HealthCheck: &elbv2.HealthCheck{Path: jsii.String(props.HealthPath), HealthyHttpCodes: jsii.String("200-399"), Interval: awscdk.Duration_Seconds(jsii.Number(30))},
	})
	svc.AttachToApplicationTargetGroup(tg)
	ln.AddTargetGroups(jsii.String("AttachTg"), &elbv2.AddApplicationTargetGroupsProps{TargetGroups: &[]elbv2.IApplicationTargetGroup{tg}})

	awscdk.NewCfnOutput(stack, jsii.String("AlbDns"), &awscdk.CfnOutputProps{Value: alb.LoadBalancerDnsName()})
	return stack
}
