package dynamodb

import (
	"context"
	"fmt"
	"os"

	appConfig "github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

//go:generate mockgen -destination=../../../testlib/mocks/dynamodb_client_mock.go -package=mocks github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform/dynamodb Client
type Client interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

func NewDynamoDbClient(conf appConfig.Config, logger platform.Logger) Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(conf.Aws.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			conf.Aws.AccessKeyId,
			conf.Aws.SecretAccessKey,
			"",
		)),
		config.WithBaseEndpoint(conf.Aws.DynamoDbLocalEndpoint),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}

	return dynamodb.NewFromConfig(cfg)
}

func GetDynamodbRegionByCountry(countryId uint16) string {
	return "us-east-2"
}

/*
package dynamodb

import (
	"context"
	"fmt"
	"os"

	appConfig "github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

//go:generate mockgen -destination=../../../testlib/mocks/dynamodb_client_mock.go -package=mocks github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform/dynamodb Client
type Client interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	PutItem(ctx context.Context, in *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, in *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, in *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, in *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, in *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, in *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(options *dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

func NewDynamoDbClient(conf appConfig.Config, logger platform.Logger) Client {
	// helper to wrap config.LoadOptionsFunc -> func(*config.LoadOptions) error
	toOpt := func(f config.LoadOptionsFunc) func(*config.LoadOptions) error {
		return func(o *config.LoadOptions) error { return f(o) }
	}

	// Build options as []func(*config.LoadOptions) error (exact type LoadDefaultConfig expects)
	opts := []func(*config.LoadOptions) error{
		toOpt(config.WithRegion(conf.Aws.Region)),
	}

	// Use static creds ONLY if explicitly provided (and not the "dummy" defaults).
	if conf.Aws.AccessKeyId != "" && conf.Aws.AccessKeyId != "dummy" &&
		conf.Aws.SecretAccessKey != "" && conf.Aws.SecretAccessKey != "dummy" {
		opts = append(opts, toOpt(config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				conf.Aws.AccessKeyId,
				conf.Aws.SecretAccessKey,
				"",
			),
		)))
	}

	// Use local/override endpoint ONLY when set (DynamoDB Local / LocalStack).
	if conf.Aws.DynamoDbLocalEndpoint != "" {
		opts = append(opts, toOpt(config.WithBaseEndpoint(conf.Aws.DynamoDbLocalEndpoint)))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}

	return dynamodb.NewFromConfig(cfg)
}

func GetDynamodbRegionByCountry(countryId uint16) string {
	return "us-east-2"
}

*/
