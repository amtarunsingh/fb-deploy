package testcontainer

import (
	"context"
	"os"
	"sync"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
	dynamodbTestcontainer "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DynamoDbLocal struct {
	Container *dynamodbTestcontainer.DynamoDBContainer
	Client    *dynamodb.Client
}

func SetupDynamoDbLocal(ctx context.Context, region string) (*DynamoDbLocal, error) {
	var (
		once          sync.Once
		initErr       error
		dynamoDbLocal *DynamoDbLocal
	)
	once.Do(func() {
		ddbContainer, err := dynamodbTestcontainer.Run(ctx,
			"amazon/dynamodb-local:3.1.0",
			dynamodbTestcontainer.WithDisableTelemetry(),
			dynamodbTestcontainer.WithSharedDB(),
			testcontainers.WithWaitStrategy(
				wait.ForListeningPort("8000/tcp").WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			initErr = err
			return
		}

		endpoint, err := ddbContainer.Endpoint(ctx, "http")
		if err != nil {
			_ = ddbContainer.Terminate(ctx)
			initErr = err
			return
		}

		// cfg, err := awsConfig.LoadDefaultConfig(
		// 	ctx,
		// 	awsConfig.WithRegion(region),
		// 	awsConfig.WithBaseEndpoint(endpoint),
		// )
		cfg, err := awsConfig.LoadDefaultConfig(
			ctx,
			awsConfig.WithRegion(region),
			awsConfig.WithBaseEndpoint(endpoint),
			awsConfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider("test", "test", ""),
			),
		)

		if err != nil {
			initErr = err
			return
		}

		dynamoDbLocal = &DynamoDbLocal{
			Container: ddbContainer,
			Client:    dynamodb.NewFromConfig(cfg),
		}
	})

	if initErr != nil {
		os.Exit(1)
	}

	return dynamoDbLocal, nil
}
