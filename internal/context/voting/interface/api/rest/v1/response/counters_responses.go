package response

import "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/entity"

type CountersGroup struct {
	IncomingYes uint32 `json:"incoming_yes" doc:"Incoming yes votes count"`
	IncomingNo  uint32 `json:"incoming_no" doc:"Incoming no votes count"`
	OutgoingYes uint32 `json:"outgoing_yes" doc:"Outgoing yes votes count"`
	OutgoingNo  uint32 `json:"outgoing_no" doc:"Outgoing no votes count"`
}

type LifetimeCountersGetResponse struct {
	Body CountersGroup
}

func CreateLifetimeCountersGetResponseFromCountersGroup(counters entity.CountersGroup) *LifetimeCountersGetResponse {
	resp := &LifetimeCountersGetResponse{
		Body: CountersGroup{
			IncomingYes: counters.IncomingYes,
			IncomingNo:  counters.IncomingNo,
			OutgoingYes: counters.OutgoingYes,
			OutgoingNo:  counters.OutgoingNo,
		},
	}
	return resp
}

type HourlyCountersGetResponse struct {
	Body map[uint8]CountersGroup
}

func CreateHourlyCountersGetResponseFromCountersGroup(counters map[uint8]*entity.CountersGroup) *HourlyCountersGetResponse {
	resp := &HourlyCountersGetResponse{
		Body: map[uint8]CountersGroup{},
	}

	for group, counter := range counters {
		resp.Body[group] = CountersGroup{
			IncomingYes: counter.IncomingYes,
			IncomingNo:  counter.IncomingNo,
			OutgoingYes: counter.OutgoingYes,
			OutgoingNo:  counter.OutgoingNo,
		}
	}

	return resp
}
