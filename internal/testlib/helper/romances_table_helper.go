package helper

import (
	"context"
	"errors"
	"fmt"
	infraDynamodb "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/infrastructure/persistence"
	platformDynamodb "github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform/dynamodb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"time"
)

type RomancesTableHelper struct {
	ddbClient platformDynamodb.Client
}

func NewRomancesTableHelper(client platformDynamodb.Client) (*RomancesTableHelper, error) {
	return &RomancesTableHelper{
		ddbClient: client,
	}, nil
}

func (c *RomancesTableHelper) CreateRomancesTable() error {
	ctx := context.Background()
	table := aws.String(infraDynamodb.RomancesTableName)

	_, err := c.ddbClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: table,
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{AttributeName: aws.String(infraDynamodb.PkUserIdAttrName), AttributeType: ddbtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String(infraDynamodb.SkUserIdAttrName), AttributeType: ddbtypes.ScalarAttributeTypeS},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{AttributeName: aws.String(infraDynamodb.PkUserIdAttrName), KeyType: ddbtypes.KeyTypeHash},
			{AttributeName: aws.String(infraDynamodb.SkUserIdAttrName), KeyType: ddbtypes.KeyTypeRange},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	})

	var condCheckErr *ddbtypes.ResourceInUseException
	if err != nil && !errors.As(err, &condCheckErr) {
		return err
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		out, err := c.ddbClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: table})
		if err == nil && out.Table != nil && out.Table.TableStatus == ddbtypes.TableStatusActive {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("table %s not ACTIVE in time", *table)
}

func (c *RomancesTableHelper) GetRomanceTableRecord(
	romanceKey infraDynamodb.RomancePrimaryKey,
	dynamoDbRegion string,
) (infraDynamodb.RomanceDocumentSchema, error) {
	out, err := c.ddbClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		Key: map[string]ddbtypes.AttributeValue{
			infraDynamodb.PkUserIdAttrName: &ddbtypes.AttributeValueMemberS{Value: romanceKey.Pk.String()},
			infraDynamodb.SkUserIdAttrName: &ddbtypes.AttributeValueMemberS{Value: romanceKey.Sk.String()},
		},
		TableName:      aws.String(infraDynamodb.RomancesTableName),
		ConsistentRead: aws.Bool(true),
	}, func(o *dynamodb.Options) {
		o.Region = dynamoDbRegion
	})

	if err != nil {
		return infraDynamodb.RomanceDocumentSchema{}, err
	}

	if len(out.Item) == 0 {
		return infraDynamodb.RomanceDocumentSchema{}, nil
	}

	romanceItem := &infraDynamodb.RomanceDocumentSchema{}
	if err = attributevalue.UnmarshalMap(out.Item, romanceItem); err != nil {
		return infraDynamodb.RomanceDocumentSchema{}, err
	}
	return *romanceItem, nil
}
