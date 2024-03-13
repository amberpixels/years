package years

import "time"

// Time is a wrapper of a standard time.Time
// The time.Time is stored intentionally under the pointer for an easy use to modify it
//
// Example:
//
//	 t, _ := time.Parse("...")
//		Wrap(&t).SomeModifyingMethod() // leads to update the t
type Time struct {
	t *time.Time
}

func Wrap(v *time.Time) *Time {
	return &Time{v}
}

// TruncateToDay overrides hour, minute, second, nanosecond to zero
func (t *Time) TruncateToDay() *Time {
	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), 0, 0, 0, 0, t.t.Location())

	return t
}

// SetYear overrides year of the time
func (t *Time) SetYear(v int) *Time {
	*(t.t) = time.Date(0, t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		AddDate(v, 0, 0)

	return t
}

// SetMonth overrides month of the time
func (t *Time) SetMonth(month time.Month) *Time {
	*(t.t) = time.Date(t.t.Year(), 0, t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		AddDate(0, int(month), 0)

	return t
}

// SetDay overrides day of the time
// Note: Feb2 .SetDay(31) will lead to ~Mar2-3 (depending on days in Feb)
func (t *Time) SetDay(day int) *Time {
	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		AddDate(0, 0, day)

	return t
}

// SetHour overrides hour of the time
func (t *Time) SetHour(hour int) *Time {
	if hour < 0 || hour >= 24 {
		panic("SetMinute accepts hour to be from 0 to 23")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), 0, t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		Add(time.Duration(hour) * time.Hour)

	return t
}

// SetMinute overrides minute of the time
func (t *Time) SetMinute(minute int) *Time {
	if minute < 0 || minute >= 60 {
		panic("SetMinute accepts minute to be from 0 to 59")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), 0, t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		Add(time.Duration(minute) * time.Minute)

	return t
}

// SetSecond overrides second of the time
func (t *Time) SetSecond(second int) *Time {
	if second < 0 || second >= 60 {
		panic("SetMinute accepts second to be from 0 to 59")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), 0, t.t.Nanosecond(), t.t.Location()).
		Add(time.Duration(second) * time.Second)

	return t
}

// SetNanosecond overrides nanosecond of the time
func (t *Time) SetNanosecond(nanosecond int) *Time {
	if nanosecond < 0 || nanosecond >= 60 {
		panic("SetMinute accepts nanosecond to be from 0 to 59")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), 0, t.t.Location()).
		Add(time.Duration(nanosecond))

	return t
}

func (t *Time) Time() time.Time { return *(t.t) }
