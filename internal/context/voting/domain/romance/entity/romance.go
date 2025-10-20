package entity

import (
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/valueobject"
	sharedValueObject "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
)

type Romance struct {
	ActiveUserVote Vote
	PeerUserVote   Vote
	Version        uint32
}

func CreateEmptyRomance(voteId sharedValueObject.VoteId) Romance {
	return Romance{
		ActiveUserVote: Vote{Id: voteId},
		PeerUserVote:   Vote{Id: voteId.ToPeerVoteId()},
	}
}

func (r *Romance) IsEmpty() bool {
	return r.ActiveUserVote.VoteType == valueobject.VoteTypeEmpty && r.PeerUserVote.VoteType == valueobject.VoteTypeEmpty
}
