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
	synth := awscdk.NewBootstraplessSynthesizer(&awscdk.BootstraplessSynthesizerProps{})

	// Data stack (always creates managed resources with fixed names)
	dataStack, dataOut := NewDataStack(app, "DataStack", &DataStackProps{
		StackProps: awscdk.StackProps{Env: resolveEnv(), Synthesizer: synth},
	})

	// Service stack (no SSM; uses default VPC and internal ALB)
	_ = dataStack
	if false {
		NewServiceStack(app, "ServiceStack", &ServiceStackProps{
			StackProps:      awscdk.StackProps{Env: resolveEnv(), Synthesizer: synth},
			ServiceName:     firstNonEmpty(os.Getenv("SERVICE_NAME"), "user-votes"),
			EcrRepoName:     firstNonEmpty(os.Getenv("ECR_REPO_NAME"), "user-votes-api"),
			ImageTag:        firstNonEmpty(os.Getenv("IMAGE_TAG"), "latest"),
			ContainerPort:   8888,
			HealthcheckPath: "/health",
			Data:            dataOut, // wire constructs
		})

	}

	app.Synth(nil)
}

// --------------------------- helpers ---------------------------------------
func resolveEnv() *awscdk.Environment {
	account := firstNonEmpty(os.Getenv("CDK_DEFAULT_ACCOUNT"), os.Getenv("AWS_ACCOUNT_ID"), "000000000000")
	region := firstNonEmpty(os.Getenv("CDK_DEFAULT_REGION"), os.Getenv("AWS_REGION"), "us-east-2")
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
