package persistence

import (
	"context"
	platformDynamodb "github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform/dynamodb"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/testlib/testcontainer"
	"log"
	"os"
	"testing"
)

var (
	ddbClient platformDynamodb.Client
)

func TestMain(m *testing.M) {
	dynamoDbLocal, err := testcontainer.SetupDynamoDbLocal(context.Background(), "us-east-2")
	if err != nil {
		log.Fatalf("failed to run dynamodb: %v", err)
	}
	ddbClient = dynamoDbLocal.Client

	code := m.Run()
	os.Exit(code)
}
