package operation

import (
	"context"
	"errors"
	"fmt"
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	countersRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/repository"
	romanceDomain "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/entity"
	romancesRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/repository"
	romancesValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
)

type ChangeUserVoteOperation struct {
	romancesRepository romancesRepo.RomancesRepository
	countersRepository countersRepo.CountersRepository
	logger             platform.Logger
}

func NewChangeUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
	countersRepository countersRepo.CountersRepository,
	logger platform.Logger,
) ChangeUserVoteOperation {
	return ChangeUserVoteOperation{
		romancesRepository: romancesRepository,
		countersRepository: countersRepository,
		logger:             logger,
	}
}

func (r *ChangeUserVoteOperation) Run(
	ctx context.Context,
	voteId sharedValueObject.VoteId,
	newVoteType romancesValueObject.VoteType,
) (entity.Vote, error) {
	tries := 0

	getRomanceOperation := NewGetRomanceOperation(r.romancesRepository)
	for {
		romance, err := getRomanceOperation.Run(ctx, voteId)
		if err != nil {
			r.logger.Error(fmt.Sprintf("GetRomance error: %+v", err))
			return entity.Vote{}, err
		}

		if !isVoteTypeCanBeChanged(romance.ActiveUserVote, newVoteType) {
			return entity.Vote{}, romanceDomain.NewChangingVoteTypeError(romance.ActiveUserVote.VoteType, newVoteType)
		}

		if newVoteType == romance.ActiveUserVote.VoteType {
			return entity.Vote{}, romanceDomain.ErrVoteDuplicate
		}

		romance, err = r.romancesRepository.ChangeActiveUserVoteTypeInRomance(
			ctx,
			romance,
			newVoteType,
		)

		if err != nil {
			if errors.Is(err, romanceDomain.ErrVersionConflict) && tries < config.DynamoDbVersionConflictRetriesCount {
				tries += 1
				continue
			}
			r.logger.Error(fmt.Sprintf("ChangeActiveUserVoteTypeInRomance error: %+v", err))
			return entity.Vote{}, err
		}

		return romance.ActiveUserVote, nil
	}
}

func isVoteTypeCanBeChanged(oldVote entity.Vote, newVoteType romancesValueObject.VoteType) bool {
	allowedVoteChangeOperations := map[romancesValueObject.VoteType][]romancesValueObject.VoteType{
		romancesValueObject.VoteTypeEmpty: {
			romancesValueObject.VoteTypeNo,
			romancesValueObject.VoteTypeYes,
			romancesValueObject.VoteTypeCrush,
			romancesValueObject.VoteTypeCompliment,
		},
		romancesValueObject.VoteTypeNo: {
			romancesValueObject.VoteTypeYes,
			romancesValueObject.VoteTypeCrush,
			romancesValueObject.VoteTypeCompliment,
		},
		romancesValueObject.VoteTypeYes: {
			romancesValueObject.VoteTypeCrush,
			romancesValueObject.VoteTypeCompliment,
		},
		romancesValueObject.VoteTypeCrush:      {},
		romancesValueObject.VoteTypeCompliment: {},
	}

	vals, ok := allowedVoteChangeOperations[oldVote.VoteType]
	if !ok {
		return false
	}

	for _, v := range vals {
		if v == newVoteType {
			return true
		}
	}
	return false
}
