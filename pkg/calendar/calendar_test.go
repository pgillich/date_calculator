package calendar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CalendarTestSuite struct {
	suite.Suite
}

func TestCalendarTestSuite(t *testing.T) {
	suite.Run(t, new(CalendarTestSuite))
}

func (s *CalendarTestSuite) TestNewCalendar() {
	testCases := []struct {
		name            string
		config          Config
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
