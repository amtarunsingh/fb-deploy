package operation

import (
	"context"
	"errors"
	"fmt"
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	countersRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/repository"
	romanceDomain "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance"
	romancesRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/repository"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
)

type DeleteUserVoteOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	countersRepository countersRepo.CountersRepository
	logger             platform.Logger
}

func NewDeleteUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
	countersRepository countersRepo.CountersRepository,
	logger platform.Logger,
) DeleteUserVoteOperation {
	return DeleteUserVoteOperation{
		romancesRepository: romancesRepository,
		countersRepository: countersRepository,
		logger:             logger,
	}
}

func (r *DeleteUserVoteOperation) Run(ctx context.Context, voteId sharedValueObject.VoteId) error {
	tries := 0

	getRomanceOperation := NewGetRomanceOperation(r.romancesRepository)
	for {
		romance, err := getRomanceOperation.Run(ctx, voteId)
		if err != nil {
			r.logger.Error(fmt.Sprintf("GetRomance error: %+v", err))
			return err
		}

		err = r.romancesRepository.DeleteActiveUserVoteFromRomance(ctx, romance)

		if err != nil {
			if errors.Is(err, romanceDomain.ErrVersionConflict) && tries < config.DynamoDbVersionConflictRetriesCount {
				tries += 1
				continue
			}
			r.logger.Error(fmt.Sprintf("DeleteUserVoteFromRomance error: %+v", err))
			return err
		}

		return nil
	}
}
