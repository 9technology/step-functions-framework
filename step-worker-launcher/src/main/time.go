package main

import (
	"fmt"
	"time"
)

func getCurrentTime() (t time.Time) {
	t = time.Now()
	return t
}

func getSydneyTime(t time.Time) (tSyd time.Time) {
	syd, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		fmt.Println("Error loading Sydney time %s", err)
	}
	tSyd = t.In(syd)
	return tSyd
}

func getSwlDateSyd(daysEarlier int) (swldate string) {
	t := getCurrentTime()
	tSyd := getSydneyTime(t)
	swlTime := tSyd.AddDate(0, 0, -daysEarlier)
	swldate = fmt.Sprintf("%d_%02d_%02d_%02d_%02d_%02d_%09d", swlTime.Year(), swlTime.Month(),
		swlTime.Day(), swlTime.Hour(), swlTime.Minute(), swlTime.Second(), swlTime.Nanosecond())
	return swldate
}

func getUniqueExecutionName() (execname string) {
	execname = getSwlDateSyd(0)
	return execname
}
