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
	todayBeginsAt := calendar.calculateDayTime(submitAt, calendar.config.WorkBegins)
	todayEndsAt := calendar.calculateDayTime(submitAt, calendar.config.WorkEnds)

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

	todayWorkDurationMax := todayEndsAt.Sub(submitAt)

	if duration < todayWorkDurationMax {
		return submitAt.Add(duration), nil
	}

	duration -= todayWorkDurationMax
	durationDays := int(duration / calendar.config.dailyWorkDuration)
	durationRemainedLast := duration % calendar.config.dailyWorkDuration
	lastDayBeginsAt := calendar.appendWorkDays(todayBeginsAt, durationDays+1)

	return lastDayBeginsAt.Add(durationRemainedLast), nil
}

func (calendar *Calendar) appendWorkDays(todayBeginsAt time.Time, durationDays int) time.Time {
	lastWorkday := time.Weekday(int(calendar.config.FirstWorkday) + calendar.config.WorkdaysInWeek - 1)
	lastDayBeginsAt := todayBeginsAt

	thisWeekDurationMax := int(lastWorkday - lastDayBeginsAt.Weekday())

	if durationDays <= thisWeekDurationMax {
		return lastDayBeginsAt.Add(time.Duration(durationDays*hoursPerDay) * time.Hour)
	}

	lastDayBeginsAt = lastDayBeginsAt.Add(
		time.Duration((thisWeekDurationMax+daysPerWeek-calendar.config.WorkdaysInWeek)*hoursPerDay) * time.Hour)

	durationDays -= thisWeekDurationMax

	durationWeeks := durationDays / calendar.config.WorkdaysInWeek
	durationRemaindedDays := durationDays % calendar.config.WorkdaysInWeek
	lastDayBeginsAt = lastDayBeginsAt.Add(time.Duration(7*durationWeeks*hoursPerDay) * time.Hour)

	return lastDayBeginsAt.Add(time.Duration(durationRemaindedDays*hoursPerDay) * time.Hour)
}

func (calendar *Calendar) formatTime(at time.Time) string {
	return at.Format(calendar.config.TimeFormat)
}

func (calendar *Calendar) calculateDayTime(submitAt time.Time, fromMidnight time.Duration) time.Time {
	return time.Date(
		submitAt.Year(),
		submitAt.Month(),
		submitAt.Day(),
		0, 0, 0, 0,
		submitAt.Location(),
	).Add(fromMidnight)
}
