package application

import (
	"context"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/operation"
	counterEntity "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/entity"
	countersValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/valueobject"
	romanceEntity "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/entity"
	romancesValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/interface/api/rest/v1/command"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/interface/api/rest/v1/query"
)

type VotingService struct {
	addUserVoteOperation         operation.AddUserVoteOperation
	deleteUserVoteOperation      operation.DeleteUserVoteOperation
	getUserVoteOperation         operation.GetUserVoteOperation
	changeUserVoteOperation      operation.ChangeUserVoteOperation
	getRomanceOperation          operation.GetRomanceOperation
	deleteRomanceOperation       operation.DeleteRomanceOperation
	deleteRomancesOperation      operation.DeleteRomancesOperation
	getLifetimeCountersOperation operation.GetLifetimeCountersOperation
	getHourlyCountersOperation   operation.GetHourlyCountersOperation
}

func NewVotingService(
	addUserVoteOperation operation.AddUserVoteOperation,
	getUserVoteOperation operation.GetUserVoteOperation,
	deleteUserVoteOperation operation.DeleteUserVoteOperation,
	changeUserVoteOperation operation.ChangeUserVoteOperation,
	getRomanceOperation operation.GetRomanceOperation,
	deleteRomanceOperation operation.DeleteRomanceOperation,
	deleteRomancesOperation operation.DeleteRomancesOperation,
	getLifetimeCountersOperation operation.GetLifetimeCountersOperation,
	getHourlyCountersOperation operation.GetHourlyCountersOperation,
) VotingService {
	return VotingService{
		addUserVoteOperation:         addUserVoteOperation,
		getUserVoteOperation:         getUserVoteOperation,
		deleteUserVoteOperation:      deleteUserVoteOperation,
		changeUserVoteOperation:      changeUserVoteOperation,
		getRomanceOperation:          getRomanceOperation,
		deleteRomanceOperation:       deleteRomanceOperation,
		deleteRomancesOperation:      deleteRomancesOperation,
		getLifetimeCountersOperation: getLifetimeCountersOperation,
		getHourlyCountersOperation:   getHourlyCountersOperation,
	}
}

func (v *VotingService) AddUserVote(ctx context.Context, command command.VoteAdd) (romanceEntity.Vote, error) {
	voteId, err := sharedValueObject.NewVoteId(
		command.CountryId,
		command.Body.ActiveUserId,
		command.Body.PeerId,
	)
	if err != nil {
		return romanceEntity.Vote{}, err
	}
	return v.addUserVoteOperation.Run(ctx, voteId, romancesValueObject.VoteType(command.Body.VoteType), command.Body.VotedAt)
}

func (v *VotingService) GetUserVote(ctx context.Context, get query.VoteGet) (romanceEntity.Vote, error) {
	voteId, err := sharedValueObject.NewVoteId(
		get.CountryId,
		get.ActiveUserId,
		get.PeerId,
	)
	if err != nil {
		return romanceEntity.Vote{}, err
	}
	return v.getUserVoteOperation.Run(ctx, voteId)
}

func (v *VotingService) DeleteUserVote(ctx context.Context, command command.DeleteVote) error {
	voteId, err := sharedValueObject.NewVoteId(
		command.CountryId,
		command.ActiveUserId,
		command.PeerId,
	)
	if err != nil {
		return err
	}
	return v.deleteUserVoteOperation.Run(ctx, voteId)
}

func (v *VotingService) ChangeUserVote(ctx context.Context, command command.ChangeVoteType) (romanceEntity.Vote, error) {
	voteId, err := sharedValueObject.NewVoteId(
		command.CountryId,
		command.ActiveUserId,
		command.PeerId,
	)
	if err != nil {
		return romanceEntity.Vote{}, err
	}
	return v.changeUserVoteOperation.Run(ctx, voteId, romancesValueObject.VoteType(command.Body.NewType))
}

func (v *VotingService) GetRomance(ctx context.Context, get query.RomanceGet) (romanceEntity.Romance, error) {
	voteId, err := sharedValueObject.NewVoteId(
		get.CountryId,
		get.ActiveUserId,
		get.PeerId,
	)
	if err != nil {
		return romanceEntity.Romance{}, err
	}
	return v.getRomanceOperation.Run(ctx, voteId)
}

func (v *VotingService) DeleteRomance(ctx context.Context, command command.DeleteRomance) error {
	voteId, err := sharedValueObject.NewVoteId(
		command.CountryId,
		command.ActiveUserId,
		command.PeerId,
	)
	if err != nil {
		return err
	}
	return v.deleteRomanceOperation.Run(ctx, voteId)
}

func (v *VotingService) DeleteRomances(ctx context.Context, command command.DeleteRomances) error {
	userKey, err := sharedValueObject.NewActiveUserKey(
		command.CountryId,
		command.ActiveUserId,
	)
	if err != nil {
		return err
	}
	return v.deleteRomancesOperation.Run(ctx, userKey)
}

func (v *VotingService) GetLifetimeCounters(ctx context.Context, query query.LifetimeCountersGet) (counterEntity.CountersGroup, error) {
	activeUserKey, err := sharedValueObject.NewActiveUserKey(
		query.CountryId,
		query.ActiveUserId,
	)
	if err != nil {
		return counterEntity.CountersGroup{}, err
	}
	return v.getLifetimeCountersOperation.Run(ctx, activeUserKey)
}

func (v *VotingService) GetHourlyCounters(ctx context.Context, query query.HourlyCountersGet) (map[uint8]*counterEntity.CountersGroup, error) {
	activeUserKey, err := sharedValueObject.NewActiveUserKey(
		query.CountryId,
		query.ActiveUserId,
	)
	if err != nil {
		return map[uint8]*counterEntity.CountersGroup{}, err
	}

	hoursOffsetGroups, err := countersValueObject.NewHoursOffsetGroups(query.HoursOffsetGroups)
	if err != nil {
		return map[uint8]*counterEntity.CountersGroup{}, err
	}
	return v.getHourlyCountersOperation.Run(ctx, activeUserKey, hoursOffsetGroups)
}
