package operation

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/entity"
	romancesRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/repository"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type GetUserVoteOperation struct {
	romancesRepository romancesRepo.RomancesRepository
}

func NewGetUserVoteOperation(
	romancesRepository romancesRepo.RomancesRepository,
) GetUserVoteOperation {
	return GetUserVoteOperation{
		romancesRepository: romancesRepository,
	}
}

func (r *GetUserVoteOperation) Run(ctx context.Context, voteId sharedValueObject.VoteId) (entity.Vote, error) {
	getRomanceOperation := NewGetRomanceOperation(r.romancesRepository)
	romance, err := getRomanceOperation.Run(ctx, voteId)
	if err != nil {
		return entity.Vote{}, err
	}

	return romance.ActiveUserVote, nil
}
