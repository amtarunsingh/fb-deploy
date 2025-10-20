package amazon_sns

import (
	"context"
	"fmt"
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/messaging"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sns"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"os"
)

type SnsPublisher struct {
	pub    *sns.Publisher
	logger platform.Logger
}

func NewSnsPublisher(config config.Config, logger platform.Logger) *SnsPublisher {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(config.Aws.Region),
		awsConfig.WithBaseEndpoint(config.Aws.SnsLocalEndpoint),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}

	pub, err := sns.NewPublisher(
		sns.PublisherConfig{
			AWSConfig: awsCfg,
			TopicResolver: TopicResolver{
				config: config,
			},
		},
		watermill.NewCaptureLogger(),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Unable to load SDK config, %v", err))
		os.Exit(1)
	}

	return &SnsPublisher{
		pub: pub,
	}
}

func (p SnsPublisher) Publish(topic messaging.Topic, m messaging.Message) error {
	wm := watermillMessage.NewMessage(m.GetId().String(), watermillMessage.Payload(m.GetPayload()))
	err := p.pub.Publish(string(topic), wm)
	if err != nil {
		return err
	}
	return nil
}
