package repository

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/entity"
	romancesValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"time"
)

//go:generate mockgen -destination=../../../../../testlib/mocks/romances_repository_mock.go -package=mocks github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/repository RomancesRepository
type RomancesRepository interface {
	GetRomance(ctx context.Context, voteId sharedValueObject.VoteId) (entity.Romance, error)
	DeleteRomance(ctx context.Context, voteId sharedValueObject.VoteId) error
	AddActiveUserVoteToRomance(
		ctx context.Context,
		romance entity.Romance,
		voteType romancesValueObject.VoteType,
		votedAt time.Time,
	) (entity.Romance, error)
	ChangeActiveUserVoteTypeInRomance(
		ctx context.Context,
		romance entity.Romance,
		newVoteType romancesValueObject.VoteType,
	) (entity.Romance, error)
	DeleteActiveUserVoteFromRomance(ctx context.Context, romance entity.Romance) error
}
