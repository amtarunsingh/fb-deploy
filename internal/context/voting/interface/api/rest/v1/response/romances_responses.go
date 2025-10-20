package response

import (
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/entity"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/interface/api/rest/v1/contract"
)

type Romance struct {
	ActiveUserVote Vote `json:"active_user_vote" doc:"Active user vote"`
	PeerUserVote   Vote `json:"peer_vote" doc:"Peer user vote"`
}

type RomanceGetResponse struct {
	Body Romance
}

func CreateRomanceGetResponseFromVoteEntity(vote entity.Romance) *RomanceGetResponse {
	resp := &RomanceGetResponse{
		Body: Romance{
			ActiveUserVote: Vote{
				VoteType:  contract.ReadUserVoteType(vote.ActiveUserVote.VoteType),
				VotedAt:   vote.ActiveUserVote.VotedAt,
				CreatedAt: vote.ActiveUserVote.CreatedAt,
				UpdatedAt: vote.ActiveUserVote.UpdatedAt,
			},
			PeerUserVote: Vote{
				VoteType:  contract.ReadUserVoteType(vote.PeerUserVote.VoteType),
				VotedAt:   vote.PeerUserVote.VotedAt,
				CreatedAt: vote.PeerUserVote.CreatedAt,
				UpdatedAt: vote.PeerUserVote.UpdatedAt,
			},
		},
	}
	return resp
}
