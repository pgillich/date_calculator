package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pgillich/date_calculator/pkg/calendar"
)

const (
	FirstWorkdayDefault   = time.Monday
	WorkdaysInWeekDefault = 5
	WorkBeginsDefault     = 9 * time.Hour
	WorkEndsDefault       = 17 * time.Hour
)

func main() {
	_, err := calendar.NewCalendar(calendar.Config{
		FirstWorkday:   FirstWorkdayDefault,
		WorkdaysInWeek: WorkdaysInWeekDefault,
		WorkBegins:     WorkBeginsDefault,
		WorkEnds:       WorkEndsDefault,
	})
	if err != nil {
		fmt.Printf("unable to init calendar: %e\n", err)
		os.Exit(1)
	}
}
