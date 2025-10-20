package repository

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/entity"
	countersValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/valueobject"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

//go:generate mockgen -destination=../../../../../testlib/mocks/counters_repository_mock.go -package=mocks github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/repository CountersRepository
type CountersRepository interface {
	GetLifetimeCounter(
		ctx context.Context,
		activeUserKey sharedValueObject.ActiveUserKey,
	) (entity.CountersGroup, error)

	GetHourlyCounters(
		ctx context.Context,
		activeUserKey sharedValueObject.ActiveUserKey,
		hoursOffsetGroups countersValueObject.HoursOffsetGroups,
	) (map[uint8]*entity.CountersGroup, error)

	IncrYesCounters(
		ctx context.Context,
		voteId sharedValueObject.VoteId,
		counterGroup countersValueObject.CounterUpdateGroup,
	)

	IncrNoCounters(
		ctx context.Context,
		voteId sharedValueObject.VoteId,
		counterGroup countersValueObject.CounterUpdateGroup,
	)
}
