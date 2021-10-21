package calendar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CalendarTestSuite struct {
	suite.Suite
}

func imeRfc3339Unsafe(value string) time.Time {
	parsedTime, _ := time.Parse(time.RFC3339, value)

	return parsedTime
}

func TestCalendarTestSuite(t *testing.T) {
	suite.Run(t, new(CalendarTestSuite))
}

func (s *CalendarTestSuite) TestNewCalendar() {
	testCases := []struct {
		name string

		config Config

		expectedCreated bool
		expectedErr     error
	}{
		{
			name: "Too many Workdays",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 7,
				WorkBegins:     9 * time.Hour,
				WorkEnds:       17 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkdays,
		},
		{
			name: "Default",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 5,
				WorkBegins:     9 * time.Hour,
				WorkEnds:       17 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: true,
			expectedErr:     nil,
		},
		{
			name: "Negative Workdays",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: -5,
				WorkBegins:     9 * time.Hour,
				WorkEnds:       17 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkdays,
		},
		{
			name: "Max Workdays",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 6,
				WorkBegins:     9 * time.Hour,
				WorkEnds:       17 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: true,
			expectedErr:     nil,
		},

		{
			name: "Negative WorkBegins",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 5,
				WorkBegins:     -9 * time.Hour,
				WorkEnds:       17 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
		{
			name: "Negative WorkEnds",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 5,
				WorkBegins:     9 * time.Hour,
				WorkEnds:       -17 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
		{
			name: "Equal Worktime",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 5,
				WorkBegins:     9 * time.Hour,
				WorkEnds:       9 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
		{
			name: "Bigger WorkBegins",
			config: Config{
				FirstWorkday:   time.Monday,
				WorkdaysInWeek: 5,
				WorkBegins:     17 * time.Hour,
				WorkEnds:       9 * time.Hour,
				TimeFormat:     TimeFormatDefault,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		s.Run(testCase.name, func() {
			calendarTest, err := NewCalendar(testCase.config)

			s.Assert().ErrorIs(err, testCase.expectedErr)

			s.Assert().Equal(testCase.expectedCreated, calendarTest != nil)
			if calendarTest != nil {
				s.Assert().Equal(testCase.config, calendarTest.config)
			}
		})
	}
}

func (s *CalendarTestSuite) TestCalculateDayTime() {
	calendarTest, err := NewCalendar(Config{
		FirstWorkday:   FirstWorkdayDefault,
		WorkdaysInWeek: WorkdaysInWeekDefault,
		WorkBegins:     WorkBeginsDefault,
		WorkEnds:       WorkEndsDefault,
		TimeFormat:     TimeFormatDefault,
	})
	s.Assert().NoError(err)

	testCases := []struct {
		name string

		submitAt time.Time

		expectedWorkBeginsAt time.Time
	}{
		{
			name:                 "After WorkBegins",
			submitAt:             imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			expectedWorkBeginsAt: imeRfc3339Unsafe("2021-10-13T09:00:00+04:00"),
		},
		{
			name:                 "Before WorkBegins",
			submitAt:             imeRfc3339Unsafe("2021-10-13T08:30:00+04:00"),
			expectedWorkBeginsAt: imeRfc3339Unsafe("2021-10-13T09:00:00+04:00"),
		},
		{
			name:                 "Before UTC midnight",
			submitAt:             imeRfc3339Unsafe("2021-10-13T02:30:00+04:00"),
			expectedWorkBeginsAt: imeRfc3339Unsafe("2021-10-13T09:00:00+04:00"),
		},
		{
			name:                 "After UTC midnight",
			submitAt:             imeRfc3339Unsafe("2021-10-13T22:30:00-04:00"),
			expectedWorkBeginsAt: imeRfc3339Unsafe("2021-10-13T09:00:00-04:00"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		s.Run(testCase.name, func() {
			workBeginsAt := calculateDayTime(testCase.submitAt, calendarTest.config.WorkBegins)

			s.Assert().Equal(testCase.expectedWorkBeginsAt, workBeginsAt)
		})
	}
}

type AdjustableWorkTimeTestSuite struct {
	suite.Suite
}

func TestAdjustableWorkTimeTestSuite(t *testing.T) {
	suite.Run(t, new(AdjustableWorkTimeTestSuite))
}

func (s *AdjustableWorkTimeTestSuite) TestAppends() {
	calendarTest, err := NewCalendar(Config{
		FirstWorkday:   FirstWorkdayDefault,
		WorkdaysInWeek: WorkdaysInWeekDefault,
		WorkBegins:     WorkBeginsDefault,
		WorkEnds:       WorkEndsDefault,
		TimeFormat:     TimeFormatDefault,
	})
	s.NoError(err)

	testCases := []struct {
		name string

		submitAt           time.Time
		turnaroundDuration time.Duration

		expectedAppendWeeks        *adjustableWorkTime
		expectedAppendWorkdayHours *adjustableWorkTime
		expectedAppendToday        *adjustableWorkTime
	}{
		{
			name:               "Same day resolved",
			submitAt:           imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(5.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
				HourToDuration(5.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
				HourToDuration(5.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T15:00:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "Short duration to next day",
			submitAt:           imeRfc3339Unsafe("2021-10-13T16:00:00+04:00"),
			turnaroundDuration: HourToDuration(2.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T16:00:00+04:00"),
				HourToDuration(2.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T16:00:00+04:00"),
				HourToDuration(2.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-14T10:30:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "Next day resolved",
			submitAt:           imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(10.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
				HourToDuration(10.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-14T09:30:00+04:00"),
				HourToDuration(2.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-14T12:00:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "2 days resolved",
			submitAt:           imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(16.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
				HourToDuration(16.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-15T09:30:00+04:00"),
				HourToDuration(0.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-15T10:00:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "Next Monday resolved",
			submitAt:           imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(24.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
				HourToDuration(24.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-18T09:30:00+04:00"),
				HourToDuration(0.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-18T10:00:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "Next 2nd Monday resolved",
			submitAt:           imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(64.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-20T09:30:00+04:00"),
				HourToDuration(24.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-25T09:30:00+04:00"),
				HourToDuration(0.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-25T10:00:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "Next 3rd Monday resolved",
			submitAt:           imeRfc3339Unsafe("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(104.5),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-27T09:30:00+04:00"),
				HourToDuration(24.5),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-11-01T09:30:00+04:00"),
				HourToDuration(0.5),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-11-01T10:00:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "39-hour from Monday morning until Friday afternoon",
			submitAt:           imeRfc3339Unsafe("2021-10-11T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(39),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-11T09:30:00+04:00"),
				HourToDuration(39),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-15T09:30:00+04:00"),
				HourToDuration(7),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-15T16:30:00+04:00"),
				HourToDuration(0),
			),
		},
		{
			name:               "79-hour from Monday morning until next Friday afternoon",
			submitAt:           imeRfc3339Unsafe("2021-10-11T09:30:00+04:00"),
			turnaroundDuration: HourToDuration(79),
			expectedAppendWeeks: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-18T09:30:00+04:00"),
				HourToDuration(39),
			),
			expectedAppendWorkdayHours: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-22T09:30:00+04:00"),
				HourToDuration(7),
			),
			expectedAppendToday: newAdjustableWorkTimeUnsafe(calendarTest,
				imeRfc3339Unsafe("2021-10-22T16:30:00+04:00"),
				HourToDuration(0),
			),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		s.Run(testCase.name, func() {
			adjustedWorkTime, err := calendarTest.newAdjustableWorkTime(
				testCase.submitAt,
				testCase.turnaroundDuration,
			)

			s.NoError(err, "new")

			adjustedWorkTime = adjustedWorkTime.appendWeeks()

			s.Assert().Equal(
				testCase.expectedAppendWeeks.time.Format(TimeFormatDefault),
				adjustedWorkTime.time.Format(TimeFormatDefault),
				"appendWeeks, time",
			)

			s.Assert().Equal(
				testCase.expectedAppendWeeks.adjust,
				adjustedWorkTime.adjust,
				"appendWeeks, adjust",
			)

			adjustedWorkTime = adjustedWorkTime.appendWorkdayHours()

			s.Assert().Equal(
				testCase.expectedAppendWorkdayHours.time.Format(TimeFormatDefault),
				adjustedWorkTime.time.Format(TimeFormatDefault),
				"appendWorkdayHours, time",
			)

			s.Assert().Equal(
				testCase.expectedAppendWorkdayHours.adjust,
				adjustedWorkTime.adjust,
				"appendWorkdayHours, adjust",
			)

			adjustedWorkTime = adjustedWorkTime.appendToday()

			s.Assert().Equal(
				testCase.expectedAppendToday.time.Format(TimeFormatDefault),
				adjustedWorkTime.time.Format(TimeFormatDefault),
				"appendWorkdayHours, time",
			)

			s.Assert().Equal(
				testCase.expectedAppendToday.adjust,
				adjustedWorkTime.adjust,
				"appendWorkdayHours, adjust",
			)
		})
	}
}

func newAdjustableWorkTimeUnsafe(calendar *Calendar, submitAt time.Time, adjust time.Duration) *adjustableWorkTime {
	workTime, _ := calendar.newAdjustableWorkTime(submitAt, adjust)

	return workTime
}
