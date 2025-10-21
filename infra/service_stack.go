package main

import (
	"fmt"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	awsec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	awsecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	awsecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	awslogs "github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ServiceStackProps struct {
	awscdk.StackProps

	ServiceName     string
	EcrRepoName     string
	ImageTag        string
	ContainerPort   int
	HealthcheckPath string
	Data            *DataOutputs // from DataStack
}

func NewServiceStack(scope constructs.Construct, id string, props *ServiceStackProps) awscdk.Stack {
	var sp awscdk.StackProps
	if props != nil {
		sp = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sp)

	if props.ServiceName == "" {
		props.ServiceName = "user-votes"
	}
	if props.ContainerPort == 0 {
		props.ContainerPort = 8888
	}
	if props.HealthcheckPath == "" {
		props.HealthcheckPath = "/health"
	}
	if props.ImageTag == "" {
		props.ImageTag = "latest"
	}
	if props.EcrRepoName == "" {
		panic("EcrRepoName must be provided")
	}

	// --------- Default VPC (org disallows custom VPCs) ----------
	vpc := awsec2.Vpc_FromLookup(stack, jsii.String("Vpc"), &awsec2.VpcLookupOptions{IsDefault: jsii.Bool(true)})

	// --------- Cluster / roles / task def ----------
	cluster := awsecs.NewCluster(stack, jsii.String("Cluster"), &awsecs.ClusterProps{
		Vpc:               vpc,
		ContainerInsights: jsii.Bool(true),
		ClusterName:       jsii.String(fmt.Sprintf("%s-cluster", props.ServiceName)),
	})

	taskRole := awsiam.NewRole(stack, jsii.String("TaskRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})
	execRole := awsiam.NewRole(stack, jsii.String("TaskExecRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AmazonECSTaskExecutionRolePolicy")),
		},
	})

	td := awsecs.NewFargateTaskDefinition(stack, jsii.String("TaskDef"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(256),
		MemoryLimitMiB: jsii.Number(512),
		TaskRole:       taskRole,
		ExecutionRole:  execRole,
	})

	repo := awsecr.Repository_FromRepositoryName(stack, jsii.String("Repo"), jsii.String(props.EcrRepoName))
	container := td.AddContainer(jsii.String("App"), &awsecs.ContainerDefinitionOptions{
		Image: awsecs.ContainerImage_FromEcrRepository(repo, jsii.String(props.ImageTag)),
		Logging: awsecs.LogDriver_AwsLogs(&awsecs.AwsLogDriverProps{
			StreamPrefix: jsii.String(props.ServiceName),
			LogGroup:     awslogs.NewLogGroup(stack, jsii.String("LogGroup"), &awslogs.LogGroupProps{Retention: awslogs.RetentionDays_ONE_WEEK}),
		}),
		PortMappings: &[]*awsecs.PortMapping{{ContainerPort: jsii.Number(float64(props.ContainerPort))}},
		Environment: &map[string]*string{
			"SERVICE_NAME": jsii.String(props.ServiceName),
		},
		HealthCheck: &awsecs.HealthCheck{
			Command:  jsii.Strings("CMD-SHELL", fmt.Sprintf("curl -sSf http://localhost:%d%s || exit 1", props.ContainerPort, props.HealthcheckPath)),
			Interval: awscdk.Duration_Seconds(jsii.Number(30)),
			Timeout:  awscdk.Duration_Seconds(jsii.Number(5)),
			Retries:  jsii.Number(3),
		},
	})

	// ---------------- Environment from constructs (no SSM) ----------------
	// Prefer construct-derived names; fall back to hard-coded defaults if Data is nil
	if props.Data != nil && props.Data.Counters != nil {
		container.AddEnvironment(jsii.String("DDB_COUNTERS"), props.Data.Counters.TableName())
	} else {
		container.AddEnvironment(jsii.String("DDB_COUNTERS"), jsii.String("Counters"))
	}
	if props.Data != nil && props.Data.Romances != nil {
		container.AddEnvironment(jsii.String("DDB_ROMANCES"), props.Data.Romances.TableName())
	} else {
		container.AddEnvironment(jsii.String("DDB_ROMANCES"), jsii.String("Romances"))
	}

	// Grants from constructs (no From*)
	if props.Data != nil {
		// DynamoDB permissions
		if props.Data.Counters != nil {
			props.Data.Counters.GrantReadWriteData(taskRole)
		}
		if props.Data.Romances != nil {
			props.Data.Romances.GrantReadWriteData(taskRole)
		}
		// Primary SNS/SQS
		if props.Data.DeleteRomancesFifoTopic != nil {
			props.Data.DeleteRomancesFifoTopic.GrantPublish(taskRole)
			container.AddEnvironment(jsii.String("SNS_TOPIC_ARN"), props.Data.DeleteRomancesFifoTopic.TopicArn())
		}
		if props.Data.DeleteRomancesFifoQueue != nil {
			props.Data.DeleteRomancesFifoQueue.GrantConsumeMessages(taskRole)
			container.AddEnvironment(jsii.String("SQS_QUEUE_ARN"), props.Data.DeleteRomancesFifoQueue.QueueArn())
			if qu := props.Data.DeleteRomancesFifoQueue.QueueUrl(); qu != nil {
				container.AddEnvironment(jsii.String("SQS_QUEUE_URL"), qu)
			}
			if qn := props.Data.DeleteRomancesFifoQueue.QueueName(); qn != nil {
				container.AddEnvironment(jsii.String("SQS_QUEUE_NAME"), qn)
			}
		}
		// Grouped SNS/SQS (optional)
		if props.Data.DeleteRomancesGroupFifoTopic != nil {
			props.Data.DeleteRomancesGroupFifoTopic.GrantPublish(taskRole)
			container.AddEnvironment(jsii.String("SNS_GROUP_TOPIC_ARN"), props.Data.DeleteRomancesGroupFifoTopic.TopicArn())
		}
		if props.Data.DeleteRomancesGroupFifoQueue != nil {
			props.Data.DeleteRomancesGroupFifoQueue.GrantConsumeMessages(taskRole)
			container.AddEnvironment(jsii.String("SQS_GROUP_QUEUE_ARN"), props.Data.DeleteRomancesGroupFifoQueue.QueueArn())
			if qu := props.Data.DeleteRomancesGroupFifoQueue.QueueUrl(); qu != nil {
				container.AddEnvironment(jsii.String("SQS_GROUP_QUEUE_URL"), qu)
			}
			if qn := props.Data.DeleteRomancesGroupFifoQueue.QueueName(); qn != nil {
				container.AddEnvironment(jsii.String("SQS_GROUP_QUEUE_NAME"), qn)
			}
		}
	}

	// --------- Internal ALB in default VPC public subnets (no NAT) ----------
	alb := elbv2.NewApplicationLoadBalancer(stack, jsii.String("Alb"), &elbv2.ApplicationLoadBalancerProps{
		Vpc:            vpc,
		InternetFacing: jsii.Bool(false), // INTERNAL only
		VpcSubnets:     &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
	})
	l := alb.AddListener(jsii.String("Http"), &elbv2.BaseApplicationListenerProps{
		Port: jsii.Number(80), Open: jsii.Bool(false), // don’t auto-open 0.0.0.0/0
	})

	// Service (public subnets + public IP for outbound ECR/CloudWatch)
	svc := awsecs.NewFargateService(stack, jsii.String("Service"), &awsecs.FargateServiceProps{
		Cluster:        cluster,
		TaskDefinition: td,
		ServiceName:    jsii.String(props.ServiceName),
		DesiredCount:   jsii.Number(1),
		AssignPublicIp: jsii.Bool(true),
		VpcSubnets:     &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
	})
	l.AddTargets(jsii.String("Attach"), &elbv2.AddApplicationTargetsProps{
		Targets:  &[]elbv2.IApplicationLoadBalancerTarget{svc},
		Port:     jsii.Number(props.ContainerPort),
		Protocol: elbv2.ApplicationProtocol_HTTP,
		HealthCheck: &elbv2.HealthCheck{
			Path:             jsii.String(props.HealthcheckPath),
			HealthyHttpCodes: jsii.String("200-399"),
			Interval:         awscdk.Duration_Seconds(jsii.Number(30)),
		},
	})

	// Explicitly allow inbound traffic to the service only from the ALB SG
	svc.Connections().AllowFrom(alb, awsec2.Port_Tcp(jsii.Number(float64(props.ContainerPort))), jsii.String("ALB to service"))

	awscdk.NewCfnOutput(stack, jsii.String("AlbDns"), &awscdk.CfnOutputProps{Value: alb.LoadBalancerDnsName()})
	return stack
}
