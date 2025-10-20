package response

import (
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/entity"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/interface/api/rest/v1/contract"
	"time"
)

type Vote struct {
	VoteType  contract.ReadUserVoteType `json:"vote_type"`
	VotedAt   *time.Time                `json:"voted_at" doc:"Vote time"`
	CreatedAt *time.Time                `json:"created_at" doc:"Vote creation time"`
	UpdatedAt *time.Time                `json:"updated_at" doc:"Vote update time"`
}

type VoteGetResponse struct {
	Body Vote
}

type VoteAddResponse struct {
	Body Vote
}

type ChangeVoteResponse struct {
	Body Vote
}

func CreateVoteGetResponseFromVoteEntity(vote entity.Vote) *VoteGetResponse {
	return &VoteGetResponse{
		Body: Vote{
			VoteType:  contract.ReadUserVoteType(vote.VoteType),
			VotedAt:   vote.VotedAt,
			CreatedAt: vote.CreatedAt,
			UpdatedAt: vote.UpdatedAt,
		},
	}
}

func CreateVoteAddResponseFromVoteEntity(vote entity.Vote) *VoteAddResponse {
	return &VoteAddResponse{
		Body: Vote{
			VoteType:  contract.ReadUserVoteType(vote.VoteType),
			VotedAt:   vote.VotedAt,
			CreatedAt: vote.CreatedAt,
			UpdatedAt: vote.UpdatedAt,
		},
	}
}

func CreateChangeVoteResponseFromVoteEntity(vote entity.Vote) *ChangeVoteResponse {
	return &ChangeVoteResponse{
		Body: Vote{
			VoteType:  contract.ReadUserVoteType(vote.VoteType),
			VotedAt:   vote.VotedAt,
			CreatedAt: vote.CreatedAt,
			UpdatedAt: vote.UpdatedAt,
		},
	}
}
