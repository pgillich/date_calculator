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
	TimeFormat     string

	dailyWorkDuration time.Duration
}

type AdjustableWorkTime struct {
	config Config
	time   time.Time
	adjust time.Duration
}

const (
	daysPerWeek = 7
	hoursPerDay = 24

	FirstWorkdayDefault   = time.Monday
	WorkdaysInWeekDefault = 5
	WorkBeginsDefault     = 9 * time.Hour
	WorkEndsDefault       = 17 * time.Hour
	TimeFormatDefault     = time.RFC3339
)

var (
	ErrInvalidWorkdays   = errors.New("invalid workdays")
	ErrInvalidWorkTime   = errors.New("invalid work datetime")
	ErrInvalidSubmitTime = errors.New("invalid submit datetime")
	ErrInvalidTimeFormat = errors.New("invalid time format")
)

type Calendar struct {
	config Config
}

func NewCalendar(config Config) (*Calendar, error) {
	if int(config.FirstWorkday)+config.WorkdaysInWeek > daysPerWeek {
		return nil, fmt.Errorf(
			"%w: %s + %d", ErrInvalidWorkdays, config.FirstWorkday.String(), config.WorkdaysInWeek,
		)
	}

	if config.WorkdaysInWeek < 1 {
		return nil, fmt.Errorf(
			"%w: %s + %d", ErrInvalidWorkdays, config.FirstWorkday.String(), config.WorkdaysInWeek,
		)
	}

	if config.WorkBegins < 0 || config.WorkBegins >= hoursPerDay*time.Hour {
		return nil, fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, config.WorkBegins.String(), config.WorkEnds.String(),
		)
	}

	if config.WorkEnds <= 0 || config.WorkEnds > hoursPerDay*time.Hour {
		return nil, fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, config.WorkBegins.String(), config.WorkEnds.String(),
		)
	}

	if config.WorkBegins == config.WorkEnds || config.WorkBegins > config.WorkEnds {
		return nil, fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, config.WorkBegins.String(), config.WorkEnds.String(),
		)
	}

	if config.TimeFormat == "" {
		return nil, fmt.Errorf(
			"%w: %s", ErrInvalidTimeFormat, config.TimeFormat,
		)
	}

	config.dailyWorkDuration = config.WorkEnds - config.WorkBegins

	return &Calendar{
		config: config,
	}, nil
}

func (calendar *Calendar) CalculateDueDate(submitAt time.Time, turnaroundDurationHour float64) (time.Time, error) {
	return calendar.calculateDueDate(submitAt, time.Duration(turnaroundDurationHour*float64(time.Hour)))
}

func (calendar *Calendar) calculateDueDate(submitAt time.Time, duration time.Duration) (time.Time, error) {
	todayBeginsAt := calculateDayTime(submitAt, calendar.config.WorkBegins)
	todayEndsAt := calculateDayTime(submitAt, calendar.config.WorkEnds)

	if submitAt.Weekday() < calendar.config.FirstWorkday ||
		submitAt.Weekday() >= calendar.config.FirstWorkday+time.Weekday(calendar.config.WorkdaysInWeek) {
		return time.Time{}, fmt.Errorf(
			"%w: %s, must be %s - %s",
			ErrInvalidSubmitTime,
			calendar.formatTime(submitAt),
			calendar.config.FirstWorkday.String(),
			(calendar.config.FirstWorkday + time.Weekday(calendar.config.WorkdaysInWeek) - 1).String(),
		)
	}

	if submitAt.Before(todayBeginsAt) || submitAt.After(todayEndsAt) {
		return time.Time{}, fmt.Errorf(
			"%w: %s, must be %s - %s",
			ErrInvalidSubmitTime,
			calendar.formatTime(submitAt),
			calendar.formatTime(todayBeginsAt),
			calendar.formatTime(todayEndsAt),
		)
	}

	dueCalculator := AdjustableWorkTime{
		config: calendar.config,
		time:   submitAt,
		adjust: duration,
	}

	return dueCalculator.appendWeeks().appendWorkdayHours().appendToday().time, nil
}

func calculateDayTime(today time.Time, fromMidnight time.Duration) time.Time {
	return time.Date(
		today.Year(),
		today.Month(),
		today.Day(),
		0, 0, 0, 0,
		today.Location(),
	).Add(fromMidnight)
}

func (workTime *AdjustableWorkTime) appendWeeks() *AdjustableWorkTime {
	if workTime.adjust == 0 {
		return workTime
	}

	durationWeek := time.Duration(workTime.config.WorkdaysInWeek) * workTime.config.dailyWorkDuration
	weeks := int(workTime.adjust / durationWeek)
	adjustRemained := workTime.adjust % durationWeek

	workTime.time = workTime.time.Add(time.Duration(hoursPerDay*daysPerWeek*weeks) * time.Hour)
	workTime.adjust = adjustRemained

	return workTime
}

func (workTime *AdjustableWorkTime) appendWorkdayHours() *AdjustableWorkTime {
	if workTime.adjust == 0 {
		return workTime
	}

	workTime.appendWeeks()

	lastWorkday := time.Weekday(int(workTime.config.FirstWorkday) + workTime.config.WorkdaysInWeek - 1)

	for workTime.adjust >= workTime.config.dailyWorkDuration {
		if workTime.time.Weekday() == lastWorkday {
			weekendAdd := time.Duration((daysPerWeek-workTime.config.WorkdaysInWeek+1)*hoursPerDay) * time.Hour
			workTime.time = workTime.time.Add(weekendAdd)
			workTime.adjust -= workTime.config.dailyWorkDuration
		} else {
			workTime.time = workTime.time.Add(hoursPerDay * time.Hour)
			workTime.adjust -= workTime.config.dailyWorkDuration
		}
	}

	return workTime
}

func (workTime *AdjustableWorkTime) appendToday() *AdjustableWorkTime {
	if workTime.adjust == 0 {
		return workTime
	}

	workTime.appendWorkdayHours()

	todayEndsAt := calculateDayTime(workTime.time, workTime.config.WorkEnds)
	todayWorkDurationMax := todayEndsAt.Sub(workTime.time)

	if workTime.adjust >= todayWorkDurationMax {
		workTime.adjust += workTime.config.dailyWorkDuration
		workTime.appendWorkdayHours()

		workTime.adjust -= workTime.config.dailyWorkDuration
	}

	workTime.time = workTime.time.Add(workTime.adjust)
	workTime.adjust = 0

	return workTime
}

func (calendar *Calendar) formatTime(at time.Time) string {
	return at.Format(calendar.config.TimeFormat)
}

func HourToDuration(hour float64) time.Duration {
	return time.Duration(hour * float64(time.Hour))
}
