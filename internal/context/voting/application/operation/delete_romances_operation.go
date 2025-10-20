package operation

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/messaging/message"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/messaging"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
)

const DeleteRomancesTopic = messaging.Topic("delete-romances")

type DeleteRomancesOperation struct {
	publisher messaging.Publisher
	logger    platform.Logger
}

func NewDeleteRomancesOperation(
	publisher messaging.Publisher,
	logger platform.Logger,
) DeleteRomancesOperation {
	return DeleteRomancesOperation{
		publisher: publisher,
		logger:    logger,
	}
}

func (r *DeleteRomancesOperation) Run(ctx context.Context, userKey sharedValueObject.ActiveUserKey) error {
	r.logger.Debug("Publishing new DeleteRomancesMessage message")
	return r.publisher.Publish(DeleteRomancesTopic, message.NewDeleteRomancesMessage(userKey))
}
