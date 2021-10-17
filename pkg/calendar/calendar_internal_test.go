package calendar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CalendarTestSuite struct {
	suite.Suite
}

func parseTimeRfc3339(value string) time.Time {
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
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    7,
				WorkBegins:        9 * time.Hour,
				WorkEnds:          17 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (17 - 9) * time.Hour,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkdays,
		},
		{
			name: "Default",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    5,
				WorkBegins:        9 * time.Hour,
				WorkEnds:          17 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (17 - 9) * time.Hour,
			},
			expectedCreated: true,
			expectedErr:     nil,
		},
		{
			name: "Negative Workdays",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    -5,
				WorkBegins:        9 * time.Hour,
				WorkEnds:          17 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (17 - 9) * time.Hour,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkdays,
		},
		{
			name: "Max Workdays",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    6,
				WorkBegins:        9 * time.Hour,
				WorkEnds:          17 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (17 - 9) * time.Hour,
			},
			expectedCreated: true,
			expectedErr:     nil,
		},

		{
			name: "Negative WorkBegins",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    5,
				WorkBegins:        -9 * time.Hour,
				WorkEnds:          17 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (17 + 9) * time.Hour,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
		{
			name: "Negative WorkEnds",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    5,
				WorkBegins:        9 * time.Hour,
				WorkEnds:          -17 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (-17 - 9) * time.Hour,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
		{
			name: "Equal Worktime",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    5,
				WorkBegins:        9 * time.Hour,
				WorkEnds:          9 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: 0,
			},
			expectedCreated: false,
			expectedErr:     ErrInvalidWorkTime,
		},
		{
			name: "Bigger WorkBegins",
			config: Config{
				FirstWorkday:      time.Monday,
				WorkdaysInWeek:    5,
				WorkBegins:        17 * time.Hour,
				WorkEnds:          9 * time.Hour,
				TimeFormat:        TimeFormatDefault,
				dailyWorkDuration: (9 - 17) * time.Hour,
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
	//nolint:exhaustivestruct // do not check missing private member setting
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
			submitAt:             parseTimeRfc3339("2021-10-13T09:30:00+04:00"),
			expectedWorkBeginsAt: parseTimeRfc3339("2021-10-13T09:00:00+04:00"),
		},
		{
			name:                 "Before WorkBegins",
			submitAt:             parseTimeRfc3339("2021-10-13T08:30:00+04:00"),
			expectedWorkBeginsAt: parseTimeRfc3339("2021-10-13T09:00:00+04:00"),
		},
		{
			name:                 "Before UTC midnight",
			submitAt:             parseTimeRfc3339("2021-10-13T02:30:00+04:00"),
			expectedWorkBeginsAt: parseTimeRfc3339("2021-10-13T09:00:00+04:00"),
		},
		{
			name:                 "After UTC midnight",
			submitAt:             parseTimeRfc3339("2021-10-13T22:30:00-04:00"),
			expectedWorkBeginsAt: parseTimeRfc3339("2021-10-13T09:00:00-04:00"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		s.Run(testCase.name, func() {
			workBeginsAt := calendarTest.calculateDayTime(testCase.submitAt, calendarTest.config.WorkBegins)

			s.Assert().Equal(testCase.expectedWorkBeginsAt, workBeginsAt)
		})
	}
}
