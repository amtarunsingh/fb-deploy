package main

import (
	"log"
	"os"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()
	log.SetFlags(0)

	app := awscdk.NewApp(nil)

	// Presence-of-values: if provided we import; otherwise we create.
	countersName := os.Getenv("COUNTERS_TABLE_NAME")
	romancesName := os.Getenv("ROMANCES_TABLE_NAME")
	topicArn := os.Getenv("DELETE_ROMANCES_TOPIC_ARN")
	queueArn := os.Getenv("DELETE_ROMANCES_QUEUE_ARN")
	queueUrl := os.Getenv("DELETE_ROMANCES_QUEUE_URL")

	// Optional app params
	ecrRepo := firstNonEmpty(os.Getenv("ECR_REPO_NAME"), "")
	imageTag := firstNonEmpty(os.Getenv("IMAGE_TAG"), "latest")
	serviceName := firstNonEmpty(os.Getenv("SERVICE_NAME"), "user-votes")

	deployService := os.Getenv("DEPLOY_SERVICE")
	if deployService == "" {
		deployService = "true"
	}

	synth := awscdk.NewBootstraplessSynthesizer(&awscdk.BootstraplessSynthesizerProps{})

	dataProps := &DataStackProps{
		StackProps:        awscdk.StackProps{Env: resolveEnv(), Synthesizer: synth},
		CountersTableName: countersName,
		RomancesTableName: romancesName,
		TopicArn:          topicArn,
		QueueArn:          queueArn,
		QueueUrl:          queueUrl,
	}
	_, _ = NewDataStack(app, "DataStack", dataProps)

	// Service stack: pass known names — if we created, they are "Counters"/"Romances".
	cn := firstNonEmpty(countersName, "Counters")
	rn := firstNonEmpty(romancesName, "Romances")

	if deployService == "true" {
		svcProps := &ServiceStackProps{
			StackProps:        awscdk.StackProps{Env: resolveEnv(), Synthesizer: synth},
			ServiceName:       serviceName,
			EcrRepoName:       ecrRepo,
			ImageTag:          imageTag,
			ContainerPort:     8888,
			HealthPath:        "/health",
			DesiredCount:      1,
			InternalOnly:      true,
			CountersTableName: cn,
			RomancesTableName: rn,
		}
		NewServiceStack(app, "ServiceStack", svcProps)
	}

	app.Synth(nil)
}

// --------------------------- helpers ---------------------------------------

func resolveEnv() *awscdk.Environment {
	account := firstNonEmpty(os.Getenv("CDK_DEFAULT_ACCOUNT"), os.Getenv("AWS_ACCOUNT_ID"), "000000000000")
	region := "us-east-2"
	return &awscdk.Environment{Account: jsii.String(account), Region: jsii.String(region)}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func optionalString(v string) *string {
	if v == "" {
		return nil
	}
	return jsii.String(v)
}
