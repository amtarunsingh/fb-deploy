// ========================= data_stack.go (SIMPLE ONE-ACCOUNT) ==============
package main

import (
	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	awsdynamodb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	awssns "github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	awssnssubscriptions "github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	awssqs "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ---------------------------------------------------------------------------
// One-account, repeatable deploys — no stage logic
// Behavior:
//   • If identifiers are PROVIDED => import/reference existing resources
//   • If identifiers are MISSING => create resources and output identifiers
// This lets you repeatedly deploy on the same account: first time creates,
// later runs import by passing env/context values captured from outputs.
// ---------------------------------------------------------------------------

type DataStackProps struct {
	awscdk.StackProps

	// If provided => import; if empty => create
	CountersTableName string
	RomancesTableName string
	TopicArn          string
	QueueArn          string
	QueueUrl          string // optional when importing SQS

	// Optional: grant tables RW to a role (ECS task/Lambda)
	GrantRwToRole awsiam.IGrantable
}

type DataOutputs struct {
	Counters awsdynamodb.ITable
	Romances awsdynamodb.ITable
	Topic    awssns.ITopic
	Queue    awssqs.IQueue
}

func NewDataStack(scope constructs.Construct, id string, props *DataStackProps) (awscdk.Stack, *DataOutputs) {
	var sp awscdk.StackProps
	if props != nil {
		sp = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sp)

	// ------------------------------ DynamoDB ----------------------------------
	createDdb := (props.CountersTableName == "" || props.RomancesTableName == "")

	var counters, romances awsdynamodb.ITable
	if createDdb {
		removal := awscdk.RemovalPolicy_RETAIN
		countersTbl := awsdynamodb.NewTable(stack, jsii.String("Counters"), &awsdynamodb.TableProps{
			TableName:           jsii.String("Counters"),
			PartitionKey:        &awsdynamodb.Attribute{Name: jsii.String("u"), Type: awsdynamodb.AttributeType_STRING},
			SortKey:             &awsdynamodb.Attribute{Name: jsii.String("h"), Type: awsdynamodb.AttributeType_NUMBER},
			BillingMode:         awsdynamodb.BillingMode_PAY_PER_REQUEST,
			PointInTimeRecovery: jsii.Bool(true),
			Encryption:          awsdynamodb.TableEncryption_AWS_MANAGED,
			RemovalPolicy:       removal,
		})
		cfnCounters := countersTbl.Node().DefaultChild().(awscdk.CfnResource)
		cfnCounters.AddOverride(jsii.String("Properties.TimeToLiveSpecification"), map[string]interface{}{"Enabled": true, "AttributeName": "ttl"})
		counters = countersTbl

		romancesTbl := awsdynamodb.NewTable(stack, jsii.String("Romances"), &awsdynamodb.TableProps{
			TableName:           jsii.String("Romances"),
			PartitionKey:        &awsdynamodb.Attribute{Name: jsii.String("a"), Type: awsdynamodb.AttributeType_STRING},
			SortKey:             &awsdynamodb.Attribute{Name: jsii.String("b"), Type: awsdynamodb.AttributeType_STRING},
			BillingMode:         awsdynamodb.BillingMode_PAY_PER_REQUEST,
			PointInTimeRecovery: jsii.Bool(true),
			Encryption:          awsdynamodb.TableEncryption_AWS_MANAGED,
			RemovalPolicy:       removal,
		})
		cfnRomances := romancesTbl.Node().DefaultChild().(awscdk.CfnResource)
		cfnRomances.AddOverride(jsii.String("Properties.TimeToLiveSpecification"), map[string]interface{}{"Enabled": true, "AttributeName": "ttl"})
		romancesTbl.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
			IndexName:      jsii.String("gsiByMaxMinUser"),
			PartitionKey:   &awsdynamodb.Attribute{Name: jsii.String("b"), Type: awsdynamodb.AttributeType_STRING},
			SortKey:        &awsdynamodb.Attribute{Name: jsii.String("a"), Type: awsdynamodb.AttributeType_STRING},
			ProjectionType: awsdynamodb.ProjectionType_KEYS_ONLY,
		})
		romances = romancesTbl

		awscdk.NewCfnOutput(stack, jsii.String("CountersTableName"), &awscdk.CfnOutputProps{Value: jsii.String("Counters")})
		awscdk.NewCfnOutput(stack, jsii.String("RomancesTableName"), &awscdk.CfnOutputProps{Value: jsii.String("Romances")})
	} else {
		counters = awsdynamodb.Table_FromTableName(stack, jsii.String("CountersImport"), jsii.String(props.CountersTableName))
		romances = awsdynamodb.Table_FromTableName(stack, jsii.String("RomancesImport"), jsii.String(props.RomancesTableName))
	}

	if props.GrantRwToRole != nil {
		counters.GrantReadWriteData(props.GrantRwToRole)
		romances.GrantReadWriteData(props.GrantRwToRole)
	}

	// ------------------------------- SNS/SQS ----------------------------------
	createMsg := (props.TopicArn == "" || props.QueueArn == "")

	var topic awssns.ITopic
	var queue awssqs.IQueue

	if createMsg {
		removal := awscdk.RemovalPolicy_RETAIN
		dlq := awssqs.NewQueue(stack, jsii.String("DeleteRomancesDLQ"), &awssqs.QueueProps{RemovalPolicy: removal})
		q := awssqs.NewQueue(stack, jsii.String("DeleteRomancesQueue"), &awssqs.QueueProps{
			QueueName:         jsii.String("delete-romances-queue"),
			DeadLetterQueue:   &awssqs.DeadLetterQueue{Queue: dlq, MaxReceiveCount: jsii.Number(5)},
			VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(90)),
			RemovalPolicy:     removal,
		})
		queue = q

		t := awssns.NewTopic(stack, jsii.String("DeleteRomancesTopic"), &awssns.TopicProps{TopicName: jsii.String("delete-romances")})
		topic = t

		q.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect:     awsiam.Effect_ALLOW,
			Principals: &[]awsiam.IPrincipal{awsiam.NewServicePrincipal(jsii.String("sns.amazonaws.com"), nil)},
			Actions:    jsii.Strings("sqs:SendMessage"),
			Resources:  &[]*string{q.QueueArn()},
			Conditions: &map[string]interface{}{"ArnEquals": map[string]interface{}{"aws:SourceArn": *t.TopicArn()}},
		}))
		awssnssubscriptions.NewSqsSubscription(q, &awssnssubscriptions.SqsSubscriptionProps{RawMessageDelivery: jsii.Bool(true)}).Bind(t)

		awscdk.NewCfnOutput(stack, jsii.String("DeleteRomancesTopicArn"), &awscdk.CfnOutputProps{Value: t.TopicArn()})
		awscdk.NewCfnOutput(stack, jsii.String("DeleteRomancesQueueArn"), &awscdk.CfnOutputProps{Value: q.QueueArn()})
		awscdk.NewCfnOutput(stack, jsii.String("DeleteRomancesQueueUrl"), &awscdk.CfnOutputProps{Value: q.QueueUrl()})
	} else {
		topic = awssns.Topic_FromTopicArn(stack, jsii.String("DeleteRomancesTopicImported"), jsii.String(props.TopicArn))
		queue = awssqs.Queue_FromQueueAttributes(stack, jsii.String("DeleteRomancesQueueImported"), &awssqs.QueueAttributes{
			QueueArn: jsii.String(props.QueueArn),
			QueueUrl: optionalString(props.QueueUrl),
		})

		awssqs.NewQueuePolicy(stack, jsii.String("DeleteRomancesQueuePolicy"), &awssqs.QueuePolicyProps{Queues: &[]awssqs.IQueue{queue}}).
			Document().AddStatements(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect:     awsiam.Effect_ALLOW,
			Principals: &[]awsiam.IPrincipal{awsiam.NewServicePrincipal(jsii.String("sns.amazonaws.com"), nil)},
			Actions:    jsii.Strings("sqs:SendMessage"),
			Resources:  &[]*string{queue.QueueArn()},
			Conditions: &map[string]interface{}{"ArnEquals": map[string]interface{}{"aws:SourceArn": props.TopicArn}},
		}))

		awssnssubscriptions.NewSqsSubscription(queue, &awssnssubscriptions.SqsSubscriptionProps{RawMessageDelivery: jsii.Bool(true)}).Bind(topic)
	}

	return stack, &DataOutputs{Counters: counters, Romances: romances, Topic: topic, Queue: queue}
}
