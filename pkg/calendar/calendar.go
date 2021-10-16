package calendar

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	FirstWorkday   time.Weekday
	WorkdaysInWeek int
	WorkBegins     time.Duration
	WorkEnds       time.Duration
}

const (
	weekDays = 7
)

var ErrInvalidWorkdays = errors.New("invalid workdays")
var ErrInvalidWorkTime = errors.New("invalid work time")

type Calendar struct {
	config Config
}

func NewCalendar(config Config) (*Calendar, error) {
	if int(config.FirstWorkday)+config.WorkdaysInWeek > weekDays {
		return nil, fmt.Errorf(
			"%w: %s + %d", ErrInvalidWorkdays, config.FirstWorkday.String(), config.WorkdaysInWeek,
		)
	}

	if config.WorkdaysInWeek < 1 {
		return nil, fmt.Errorf(
			"%w: %s + %d", ErrInvalidWorkdays, config.FirstWorkday.String(), config.WorkdaysInWeek,
		)
	}

	if config.WorkBegins < 0 || config.WorkBegins >= 24*time.Hour {
		return nil, fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, config.WorkBegins.String(), config.WorkEnds.String(),
		)
	}

	if config.WorkEnds <= 0 || config.WorkEnds > 24*time.Hour {
		return nil, fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, config.WorkBegins.String(), config.WorkEnds.String(),
		)
	}

	if config.WorkBegins == config.WorkEnds || config.WorkBegins > config.WorkEnds {
		return nil, fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, config.WorkBegins.String(), config.WorkEnds.String(),
		)
	}

	return &Calendar{
		config: config,
	}, nil
}
