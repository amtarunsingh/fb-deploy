package app

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/messaging/handler"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/operation"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/messaging"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
)

type MessageProcessor struct {
	subscriber            messaging.Subscriber
	deleteRomancesHandler handler.DeleteRomancesHandler
	logger                platform.Logger
}

func NewMessageProcessor(
	subscriber messaging.Subscriber,
	deleteRomancesHandler handler.DeleteRomancesHandler,
	logger platform.Logger,
) *MessageProcessor {
	return &MessageProcessor{
		subscriber:            subscriber,
		deleteRomancesHandler: deleteRomancesHandler,
		logger:                logger,
	}
}

func (s *MessageProcessor) Start(ctx context.Context) error {
	cancel, err := messaging.Listen(ctx, s.subscriber, operation.DeleteRomancesTopic, s.deleteRomancesHandler)
	if err != nil {
		return err
	}
	defer func() {
		if err = cancel(); err != nil {
			s.logger.Error("cancel failed", "err", err)
		}
	}()

	<-ctx.Done()
	return ctx.Err()
}
