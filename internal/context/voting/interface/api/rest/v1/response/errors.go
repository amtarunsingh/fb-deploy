package response

import (
	"errors"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/app/api/response"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance"
	"net/http"
)

func ToApiError(err error) error {
	switch {
	case errors.Is(err, romance.ErrVoteNotFound):
		return NewErr404NotFound(err.Error())
	case errors.Is(err, romance.ErrVoteDuplicate):
		return NewErr400BadRequest(err.Error())
	case errors.Is(err, romance.ErrWrongVote):
		return NewErr400BadRequest(err.Error())
	default:
		return NewErr500InternalServerError("Internal error")
	}
}

func NewErr404NotFound(msg string) *response.HumaApiError {
	return &response.HumaApiError{
		Message: msg,
		Status:  http.StatusNotFound,
	}
}

func NewErr400BadRequest(msg string) *response.HumaApiError {
	return &response.HumaApiError{
		Message: msg,
		Status:  http.StatusBadRequest,
	}
}

func NewErr500InternalServerError(msg string) *response.HumaApiError {
	return &response.HumaApiError{
		Message: msg,
		Status:  http.StatusInternalServerError,
	}
}
