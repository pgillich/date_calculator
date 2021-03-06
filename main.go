package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pgillich/date_calculator/pkg/calendar"
)

func main() {
	calendarTest, err := calendar.NewCalendar(calendar.Config{
		FirstWorkday:   calendar.FirstWorkdayDefault,
		WorkdaysInWeek: calendar.WorkdaysInWeekDefault,
		WorkBegins:     calendar.WorkBeginsDefault,
		WorkEnds:       calendar.WorkEndsDefault,
		TimeFormat:     calendar.TimeFormatDefault,
	})
	if err != nil {
		fmt.Printf("unable to init calendar: %e\n", err)
		os.Exit(1)
	}

	submitAt, err := time.Parse(time.RFC3339, "2021-10-13T09:30:00+04:00")
	if err != nil {
		fmt.Printf("unable to parse datetime: %e\n", err)
		os.Exit(1)
	}

	turnaroundDuration := 9.5

	// Example calling CalculateDueDate as a member function

	resolvedAt, err := calendarTest.CalculateDueDate(submitAt, turnaroundDuration)
	if err != nil {
		fmt.Printf("unable to calculate issue resolved datetime: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(resolvedAt.Format(time.RFC3339))

	// Example calling CalculateDueDate as a generated standalone function

	calculateDueDate := calendarTest.CalculateDueDateFunc()

	resolvedAt, err = calculateDueDate(submitAt, turnaroundDuration)
	if err != nil {
		fmt.Printf("unable to calculate issue resolved datetime: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(resolvedAt.Format(time.RFC3339))
}
