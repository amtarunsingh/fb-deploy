package message

import (
	"encoding/json"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/sharedkernel/valueobject"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/messaging"
	"github.com/google/uuid"
)

type DeleteRomancesMessage struct {
	Id           uuid.UUID `json:"id"`
	ActiveUserId uuid.UUID `json:"active_user_id"`
	CountryId    uint16    `json:"country_id"`
}

func NewDeleteRomancesMessage(activeUserKey valueobject.ActiveUserKey) *DeleteRomancesMessage {
	return &DeleteRomancesMessage{
		Id:           uuid.New(),
		ActiveUserId: activeUserKey.ActiveUserId(),
		CountryId:    activeUserKey.CountryId(),
	}
}

func (m *DeleteRomancesMessage) GetId() uuid.UUID {
	return m.Id
}

func (m *DeleteRomancesMessage) GetPayload() messaging.Payload {
	payload, _ := json.Marshal(m)
	return payload
}

func (m *DeleteRomancesMessage) Load(payload messaging.Payload) error {
	return json.Unmarshal(payload, &m)
}
