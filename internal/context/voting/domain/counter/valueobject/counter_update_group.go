package valueobject

import (
	"fmt"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/timeutil"
	"time"
)

type CounterUpdateGroup struct {
	hourStartTime time.Time
}

func NewCounterUpdateGroup(eventTime time.Time) (CounterUpdateGroup, error) {
	if eventTime.IsZero() {
		return CounterUpdateGroup{}, fmt.Errorf("eventTime must not be zero")
	}

	return CounterUpdateGroup{
		hourStartTime: timeutil.HourStart(eventTime),
	}, nil
}

func (c CounterUpdateGroup) HourStartTime() time.Time {
	return c.hourStartTime
}
