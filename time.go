package years

import "time"

// Time is a wrapper of a standard time.Time
// The time.Time is stored intentionally under the pointer for an easy use to modify it
//
// Example:
//
//	 t, _ := time.Parse("...")
//		New(&t).SomeModifyingMethod() // leads to update the t
type Time struct {
	*time.Time
}

func New(v *time.Time) *Time {
	return &Time{v}
}

// TruncateToDay overrides hour, minute, second, nanosecond to zero
func (t *Time) TruncateToDay() *Time {
	truncated := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	*(t.Time) = truncated
	return t
}

// SetYear overrides year of the time
func (t *Time) SetYear(v int) *Time {
	noYear := time.Date(0, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	yearUpdated := noYear.AddDate(v, 0, 0)
	*(t.Time) = yearUpdated

	return t
}

// SetMonth overrides month of the time
func (t *Time) SetMonth(month time.Month) *Time {
	noMonth := time.Date(t.Year(), 0, t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	monthUpdated := noMonth.AddDate(0, int(month), 0)
	*(t.Time) = monthUpdated

	return t
}

// SetDay overrides day of the time
// Note: Feb2 .SetDay(31) will lead to ~Mar2-3 (depending on days in Feb)
func (t *Time) SetDay(day int) *Time {
	noDay := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	dayUpdated := noDay.AddDate(0, 0, day)
	*(t.Time) = dayUpdated

	return t
}

// SetHour overrides hour of the time
func (t *Time) SetHour(hour int) *Time {
	if hour < 0 || hour >= 24 {
		panic("SetMinute accepts hour to be from 0 to 23")
	}

	noHour := time.Date(t.Year(), t.Month(), t.Day(), 0, t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	hourUpdated := noHour.Add(time.Duration(hour) * time.Hour)
	*(t.Time) = hourUpdated

	return t
}

// SetMinute overrides minute of the time
func (t *Time) SetMinute(minute int) *Time {
	if minute < 0 || minute >= 60 {
		panic("SetMinute accepts minute to be from 0 to 59")
	}

	noMinute := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, t.Second(), t.Nanosecond(), t.Location())
	minuteUpdated := noMinute.Add(time.Duration(minute) * time.Minute)
	*(t.Time) = minuteUpdated

	return t
}

// SetSecond overrides second of the time
func (t *Time) SetSecond(second int) *Time {
	if second < 0 || second >= 60 {
		panic("SetMinute accepts second to be from 0 to 59")
	}

	noSecond := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, t.Nanosecond(), t.Location())
	secondUpdated := noSecond.Add(time.Duration(second) * time.Second)
	*(t.Time) = secondUpdated

	return t
}

// SetNanosecond overrides nanosecond of the time
func (t *Time) SetNanosecond(nanosecond int) *Time {
	if nanosecond < 0 || nanosecond >= 60 {
		panic("SetMinute accepts nanosecond to be from 0 to 59")
	}

	noNanosecond := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	nanosecondUpdated := noNanosecond.Add(time.Duration(nanosecond))
	*(t.Time) = nanosecondUpdated

	return t
}
