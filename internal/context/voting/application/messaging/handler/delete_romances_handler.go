package handler

import (
	"context"
	"fmt"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/messaging/message"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
)

type DeleteRomancesHandler struct {
	logger platform.Logger
}

func NewDeleteDeleteRomancesHandler(
	logger platform.Logger,
) DeleteRomancesHandler {
	return DeleteRomancesHandler{
		logger: logger,
	}
}

func (h DeleteRomancesHandler) Handle(ctx context.Context, message *message.DeleteRomancesMessage) error {
	h.logger.Debug(fmt.Sprintf("message DeleteRomancesMessage received: %v", message))
	return nil
}
