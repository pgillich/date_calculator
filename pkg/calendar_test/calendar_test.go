package calendar_test

import (
	"testing"
	"time"

	"github.com/pgillich/date_calculator/pkg/calendar"
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

func (s *CalendarTestSuite) TestCalculateDueDate() {
	calendarTest, err := calendar.NewCalendar(calendar.Config{
		FirstWorkday:   calendar.FirstWorkdayDefault,
		WorkdaysInWeek: calendar.WorkdaysInWeekDefault,
		WorkBegins:     calendar.WorkBeginsDefault,
		WorkEnds:       calendar.WorkEndsDefault,
		TimeFormat:     calendar.TimeFormatDefault,
	})
	s.Assert().NoError(err)

	testCases := []struct {
		name string

		submitAt           time.Time
		turnaroundDuration float64

		expectedResolvedAt time.Time
		expectedErr        error
	}{
		{
			name:               "Weekend submitAt",
			submitAt:           parseTimeRfc3339("2021-10-16T08:30:00+04:00"),
			turnaroundDuration: 9.5,
			expectedResolvedAt: time.Time{},
			expectedErr:        calendar.ErrInvalidSubmitTime,
		},
		{
			name:               "Too early submitAt",
			submitAt:           parseTimeRfc3339("2021-10-13T08:30:00+04:00"),
			turnaroundDuration: 9.5,
			expectedResolvedAt: time.Time{},
			expectedErr:        calendar.ErrInvalidSubmitTime,
		},
		{
			name:               "Too late submitAt",
			submitAt:           parseTimeRfc3339("2021-10-13T17:30:00+04:00"),
			turnaroundDuration: 9.5,
			expectedResolvedAt: time.Time{},
			expectedErr:        calendar.ErrInvalidSubmitTime,
		},
		{
			name:               "Same day resolved",
			submitAt:           parseTimeRfc3339("2021-10-13T09:30:00+04:00"),
			turnaroundDuration: 5.5,
			expectedResolvedAt: parseTimeRfc3339("2021-10-13T15:00:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "Short duration to next day",
			submitAt:           parseTimeRfc3339("2021-10-13T16:20:00+04:00"),
			turnaroundDuration: 2,
			expectedResolvedAt: parseTimeRfc3339("2021-10-14T10:20:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "Next day resolved",
			submitAt:           parseTimeRfc3339("2021-10-13T09:20:00+04:00"),
			turnaroundDuration: 10.5,
			expectedResolvedAt: parseTimeRfc3339("2021-10-14T11:50:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "2 days resolved",
			submitAt:           parseTimeRfc3339("2021-10-13T09:20:00+04:00"),
			turnaroundDuration: 16.5,
			expectedResolvedAt: parseTimeRfc3339("2021-10-15T09:50:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "Next Monday resolved",
			submitAt:           parseTimeRfc3339("2021-10-13T09:20:00+04:00"),
			turnaroundDuration: 24.5,
			expectedResolvedAt: parseTimeRfc3339("2021-10-18T09:50:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "Next 2nd Monday resolved",
			submitAt:           parseTimeRfc3339("2021-10-13T09:20:00+04:00"),
			turnaroundDuration: 64.5,
			expectedResolvedAt: parseTimeRfc3339("2021-10-25T09:50:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "Next 3rd Monday resolved",
			submitAt:           parseTimeRfc3339("2021-10-13T09:20:00+04:00"),
			turnaroundDuration: 104.5,
			expectedResolvedAt: parseTimeRfc3339("2021-11-01T09:50:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "39-hour from Monday morning until Friday afternoon",
			submitAt:           parseTimeRfc3339("2021-10-11T09:30:00+04:00"),
			turnaroundDuration: 39,
			expectedResolvedAt: parseTimeRfc3339("2021-10-15T16:30:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "71-hour from Monday morning until next Tuesday afternoon",
			submitAt:           parseTimeRfc3339("2021-10-11T09:30:00+04:00"),
			turnaroundDuration: 71,
			expectedResolvedAt: parseTimeRfc3339("2021-10-21T16:30:00+04:00"),
			expectedErr:        nil,
		},
		{
			name:               "79-hour from Monday morning until next Friday afternoon",
			submitAt:           parseTimeRfc3339("2021-10-11T09:20:00+04:00"),
			turnaroundDuration: 79,
			expectedResolvedAt: parseTimeRfc3339("2021-10-22T16:20:00+04:00"),
			expectedErr:        nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		s.Run(testCase.name, func() {
			resolvedAt, err := calendarTest.CalculateDueDate(testCase.submitAt, testCase.turnaroundDuration)

			s.Assert().ErrorIs(err, testCase.expectedErr)

			s.Assert().Equal(
				testCase.expectedResolvedAt.Format(calendar.TimeFormatDefault),
				resolvedAt.Format(calendar.TimeFormatDefault),
			)
		})
	}
}
