package domain

import (
	"fmt"
	"time"
)

type Date struct {
	year  int
	month time.Month
	day   int
}

func ParseDate(text string) (*Date, error) {
	parsed, err := time.Parse(time.DateOnly, text)
	if err != nil {
		return nil, ErrInvalidDate.WithInternal(err)
	}
	if len(text) != len(time.DateOnly) || parsed.Year() < 1 || parsed.Format(time.DateOnly) != text {
		return nil, ErrInvalidDate
	}

	return &Date{
		year:  parsed.Year(),
		month: parsed.Month(),
		day:   parsed.Day(),
	}, nil
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func (d Date) Year() int {
	return d.year
}

func (d Date) Month() time.Month {
	return d.month
}

func (d Date) Day() int {
	return d.day
}
