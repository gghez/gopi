package types

import "time"

type Date time.Time
type MonthDate time.Time

func (t Date) MarshalYAML() (interface{}, error) {
	return time.Time(t).Format("2006-01-02"), nil
}

func (t MonthDate) MarshalYAML() (interface{}, error) {
	return time.Time(t).Format("January 2006"), nil
}
