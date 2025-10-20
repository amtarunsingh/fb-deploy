package amazon_sns

import (
	"context"
	"fmt"
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/messaging"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sns"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"os"
	"strings"
)

type SnsSubscriber struct {
	wrappedSubscriber *sns.Subscriber
	logger            platform.Logger
}

type TopicResolver struct {
	config config.Config
}

func (t TopicResolver) ResolveTopic(ctx context.Context, topic string) (snsTopic sns.TopicArn, err error) {
	return sns.TopicArn(fmt.Sprintf("arn:aws:sns:%s:%s:%s", t.config.Aws.Region, "000000000000", topic)), nil
}

func NewSnsSubscriber(
	config config.Config,
	logger platform.Logger,
) *SnsSubscriber {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(config.Aws.Region),
		awsConfig.WithBaseEndpoint(config.Aws.SnsLocalEndpoint),
	)
	snsCfg := sns.SubscriberConfig{
		AWSConfig: awsCfg,
		GenerateSqsQueueName: func(ctx context.Context, topicArn sns.TopicArn) (string, error) {
			parts := strings.Split(string(topicArn), ":")
			name := parts[len(parts)-1]
			return name + "-queue", nil
		},
		TopicResolver: TopicResolver{
			config: config,
		},
	}

	sqsCfg := sqs.SubscriberConfig{
		AWSConfig: awsCfg,
	}

	subscriber, err := sns.NewSubscriber(snsCfg, sqsCfg, watermill.NewCaptureLogger())
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	return &SnsSubscriber{
		wrappedSubscriber: subscriber,
		logger:            logger,
	}
}

func (p SnsSubscriber) Subscribe(ctx context.Context, topic messaging.Topic) (<-chan messaging.BackMessage, error) {
	messages, err := p.wrappedSubscriber.Subscribe(ctx, string(topic))
	if err != nil {
		return nil, err
	}

	out := make(chan messaging.BackMessage, 64)

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-messages:
				if !ok {
					return
				}
				out <- newSnsBackMessage(m)
			}
		}
	}()

	return out, nil
}

type SnsBackMessage struct {
	wrappedMessage *message.Message
}

func newSnsBackMessage(message *message.Message) *SnsBackMessage {
	return &SnsBackMessage{
		wrappedMessage: message,
	}
}

func (bm *SnsBackMessage) GetPayload() messaging.Payload {
	return messaging.Payload(bm.wrappedMessage.Payload)
}
func (bm *SnsBackMessage) Nack() bool {
	return bm.wrappedMessage.Nack()
}
func (bm *SnsBackMessage) Ack() bool {
	return bm.wrappedMessage.Ack()
}
