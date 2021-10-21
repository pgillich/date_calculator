package calendar

import (
	"errors"
	"fmt"
	"time"
)

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

type Config struct {
	FirstWorkday   time.Weekday
	WorkdaysInWeek int
	WorkBegins     time.Duration
	WorkEnds       time.Duration
	TimeFormat     string
}

type Calendar struct {
	config Config
}

func NewCalendar(config Config) (*Calendar, error) {
	calendar := &Calendar{
		config: config,
	}

	if _, err := calendar.newAdjustableWorkTime(time.Time{}, 0); err != nil {
		return nil, err
	}

	return calendar, nil
}

func (calendar *Calendar) CalculateDueDate(submitAt time.Time, turnaroundDurationHour float64) (time.Time, error) {
	return calendar.calculateDueDate(submitAt, time.Duration(turnaroundDurationHour*float64(time.Hour)))
}

func (calendar *Calendar) CalculateDueDateFunc() func(
	submitAt time.Time, turnaroundDurationHour float64,
) (time.Time, error) {
	return func(submitAt time.Time, turnaroundDurationHour float64) (time.Time, error) {
		return calendar.CalculateDueDate(submitAt, turnaroundDurationHour)
	}
}

func (calendar *Calendar) calculateDueDate(submitAt time.Time, duration time.Duration) (time.Time, error) {
	dueCalculator, err := calendar.newAdjustableWorkTime(submitAt, duration)
	if err != nil {
		return time.Time{}, err
	}

	return dueCalculator.appendWeeks().appendWorkdayHours().appendToday().time, nil
}

type adjustableWorkTime struct {
	config Config
	time   time.Time
	adjust time.Duration
}

func (calendar *Calendar) newAdjustableWorkTime(submitAt time.Time, adjust time.Duration) (*adjustableWorkTime, error) {
	workTime := &adjustableWorkTime{
		config: calendar.config,
		time:   submitAt,
		adjust: adjust,
	}

	if err := workTime.validateConfig(); err != nil {
		return nil, err
	}

	if submitAt.IsZero() && adjust == 0 {
		return workTime, nil
	}

	if err := workTime.validateParams(); err != nil {
		return nil, err
	}

	return workTime, nil
}

func (workTime *adjustableWorkTime) validateConfig() error {
	if int(workTime.config.FirstWorkday)+workTime.config.WorkdaysInWeek > daysPerWeek {
		return fmt.Errorf(
			"%w: %s + %d", ErrInvalidWorkdays, workTime.config.FirstWorkday.String(), workTime.config.WorkdaysInWeek,
		)
	}

	if workTime.config.WorkdaysInWeek < 1 {
		return fmt.Errorf(
			"%w: %s + %d", ErrInvalidWorkdays, workTime.config.FirstWorkday.String(), workTime.config.WorkdaysInWeek,
		)
	}

	if workTime.config.WorkBegins < 0 || workTime.config.WorkBegins >= hoursPerDay*time.Hour {
		return fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, workTime.config.WorkBegins.String(), workTime.config.WorkEnds.String(),
		)
	}

	if workTime.config.WorkEnds <= 0 || workTime.config.WorkEnds > hoursPerDay*time.Hour {
		return fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, workTime.config.WorkBegins.String(), workTime.config.WorkEnds.String(),
		)
	}

	if workTime.config.WorkBegins == workTime.config.WorkEnds || workTime.config.WorkBegins > workTime.config.WorkEnds {
		return fmt.Errorf(
			"%w: %s - %s", ErrInvalidWorkTime, workTime.config.WorkBegins.String(), workTime.config.WorkEnds.String(),
		)
	}

	if workTime.config.TimeFormat == "" {
		return fmt.Errorf(
			"%w: %s", ErrInvalidTimeFormat, workTime.config.TimeFormat,
		)
	}

	return nil
}

func (workTime *adjustableWorkTime) validateParams() error {
	todayBeginsAt := calculateDayTime(workTime.time, workTime.config.WorkBegins)
	todayEndsAt := calculateDayTime(workTime.time, workTime.config.WorkEnds)

	if workTime.time.Weekday() < workTime.config.FirstWorkday ||
		workTime.time.Weekday() >= workTime.config.FirstWorkday+time.Weekday(workTime.config.WorkdaysInWeek) {
		return fmt.Errorf(
			"%w: %s, must be %s - %s",
			ErrInvalidSubmitTime,
			workTime.formatTime(workTime.time),
			workTime.config.FirstWorkday.String(),
			(workTime.config.FirstWorkday + time.Weekday(workTime.config.WorkdaysInWeek) - 1).String(),
		)
	}

	if workTime.time.Before(todayBeginsAt) || workTime.time.After(todayEndsAt) {
		return fmt.Errorf(
			"%w: %s, must be %s - %s",
			ErrInvalidSubmitTime,
			workTime.formatTime(workTime.time),
			workTime.formatTime(todayBeginsAt),
			workTime.formatTime(todayEndsAt),
		)
	}

	return nil
}

func (workTime *adjustableWorkTime) dailyWorkDuration() time.Duration {
	return workTime.config.WorkEnds - workTime.config.WorkBegins
}

func (workTime *adjustableWorkTime) appendWeeks() *adjustableWorkTime {
	if workTime.adjust == 0 {
		return workTime
	}

	durationWeek := time.Duration(workTime.config.WorkdaysInWeek) * workTime.dailyWorkDuration()
	weeks := int(workTime.adjust / durationWeek)
	adjustRemained := workTime.adjust % durationWeek

	workTime.time = workTime.time.Add(time.Duration(hoursPerDay*daysPerWeek*weeks) * time.Hour)
	workTime.adjust = adjustRemained

	return workTime
}

func (workTime *adjustableWorkTime) appendWorkdayHours() *adjustableWorkTime {
	if workTime.adjust == 0 {
		return workTime
	}

	workTime.appendWeeks()

	lastWorkday := time.Weekday(int(workTime.config.FirstWorkday) + workTime.config.WorkdaysInWeek - 1)

	for workTime.adjust >= workTime.dailyWorkDuration() {
		if workTime.time.Weekday() == lastWorkday {
			weekendAdd := time.Duration((daysPerWeek-workTime.config.WorkdaysInWeek+1)*hoursPerDay) * time.Hour
			workTime.time = workTime.time.Add(weekendAdd)
			workTime.adjust -= workTime.dailyWorkDuration()
		} else {
			workTime.time = workTime.time.Add(hoursPerDay * time.Hour)
			workTime.adjust -= workTime.dailyWorkDuration()
		}
	}

	return workTime
}

func (workTime *adjustableWorkTime) appendToday() *adjustableWorkTime {
	if workTime.adjust == 0 {
		return workTime
	}

	workTime.appendWorkdayHours()

	todayEndsAt := calculateDayTime(workTime.time, workTime.config.WorkEnds)
	todayWorkDurationMax := todayEndsAt.Sub(workTime.time)

	if workTime.adjust >= todayWorkDurationMax {
		workTime.adjust += workTime.dailyWorkDuration()
		workTime.appendWorkdayHours()

		workTime.adjust -= workTime.dailyWorkDuration()
	}

	workTime.time = workTime.time.Add(workTime.adjust)
	workTime.adjust = 0

	return workTime
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

func (workTime *adjustableWorkTime) formatTime(at time.Time) string {
	return at.Format(workTime.config.TimeFormat)
}

func HourToDuration(hour float64) time.Duration {
	return time.Duration(hour * float64(time.Hour))
}
