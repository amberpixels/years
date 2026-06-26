package years

import "time"

// MutatingTime is a wrapper around a time.Time pointer,
// providing fluent setter methods that mutate the underlying time.
//
// Example:
//
//	t, _ := time.Parse("2006-01-02 15:04:05", "2025-04-30 13:45:00")
//	Mutate(&t).SetMonth(time.April).SetYear(2021)
//	// t is now 2021-04-30 13:45:00.
//
// Use Mutate to obtain a MutatingTime for in-place modifications.
type MutatingTime struct {
	t *time.Time
}

// Mutate returns a MutatingTime for the given *time.Time.
func Mutate(v *time.Time) *MutatingTime {
	return &MutatingTime{t: v}
}

// TruncateToSecond sets the nanosecond to zero, keeping everything down to the second.
func (mt *MutatingTime) TruncateToSecond() *MutatingTime {
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), mt.t.Second(), 0,
		mt.t.Location(),
	)
	return mt
}

// TruncateToMinute sets the second and nanosecond to zero.
func (mt *MutatingTime) TruncateToMinute() *MutatingTime {
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), 0, 0,
		mt.t.Location(),
	)
	return mt
}

// TruncateToHour sets the minute, second, and nanosecond to zero.
func (mt *MutatingTime) TruncateToHour() *MutatingTime {
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), 0, 0, 0,
		mt.t.Location(),
	)
	return mt
}

// TruncateToDay sets the hour, minute, second, and nanosecond to zero.
func (mt *MutatingTime) TruncateToDay() *MutatingTime {
	*mt.t = time.Date(mt.t.Year(), mt.t.Month(), mt.t.Day(), 0, 0, 0, 0, mt.t.Location())
	return mt
}

// TruncateToWeek moves the time back to the start of its week (at 00:00:00),
// where weekStartsOn selects which weekday begins the week (e.g. time.Monday for
// ISO-8601 weeks, time.Sunday for US-style weeks).
func (mt *MutatingTime) TruncateToWeek(weekStartsOn time.Weekday) *MutatingTime {
	mt.TruncateToDay()
	offset := (int(mt.t.Weekday()) - int(weekStartsOn) + daysInWeek) % daysInWeek
	*mt.t = mt.t.AddDate(0, 0, -offset)
	return mt
}

// TruncateToMonth moves the time to the first day of its month at 00:00:00.
func (mt *MutatingTime) TruncateToMonth() *MutatingTime {
	*mt.t = time.Date(mt.t.Year(), mt.t.Month(), 1, 0, 0, 0, 0, mt.t.Location())
	return mt
}

// TruncateToYear moves the time to January 1st of its year at 00:00:00.
func (mt *MutatingTime) TruncateToYear() *MutatingTime {
	*mt.t = time.Date(mt.t.Year(), 1, 1, 0, 0, 0, 0, mt.t.Location())
	return mt
}

// SetYear sets the year to the provided value.
func (mt *MutatingTime) SetYear(year int) *MutatingTime {
	*mt.t = time.Date(
		year,
		mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), mt.t.Second(), mt.t.Nanosecond(),
		mt.t.Location(),
	)
	return mt
}

// SetMonth sets the month to the provided value.
func (mt *MutatingTime) SetMonth(month time.Month) *MutatingTime {
	*mt.t = time.Date(
		mt.t.Year(), month, mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), mt.t.Second(), mt.t.Nanosecond(),
		mt.t.Location(),
	)
	return mt
}

// SetDay sets the day of the month to the provided value.
func (mt *MutatingTime) SetDay(day int) *MutatingTime {
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), day,
		mt.t.Hour(), mt.t.Minute(), mt.t.Second(), mt.t.Nanosecond(),
		mt.t.Location(),
	)
	return mt
}

// SetHour sets the hour (0–23). Panics if out of range.
func (mt *MutatingTime) SetHour(hour int) *MutatingTime {
	if hour < 0 || hour > 23 {
		panic("SetHour accepts hour in [0,23]")
	}
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		hour, mt.t.Minute(), mt.t.Second(), mt.t.Nanosecond(),
		mt.t.Location(),
	)
	return mt
}

// SetMinute sets the minute (0–59). Panics if out of range.
func (mt *MutatingTime) SetMinute(minute int) *MutatingTime {
	if minute < 0 || minute > 59 {
		panic("SetMinute accepts minute in [0,59]")
	}
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), minute, mt.t.Second(), mt.t.Nanosecond(),
		mt.t.Location(),
	)
	return mt
}

// SetSecond sets the second (0–59). Panics if out of range.
func (mt *MutatingTime) SetSecond(second int) *MutatingTime {
	if second < 0 || second > 59 {
		panic("SetSecond accepts second in [0,59]")
	}
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), second, mt.t.Nanosecond(),
		mt.t.Location(),
	)
	return mt
}

// SetMillisecond sets the millisecond (0–999) by overriding the nanosecond field.
// Panics if out of range.
func (mt *MutatingTime) SetMillisecond(ms int) *MutatingTime {
	if ms < 0 || ms > 999 {
		panic("SetMillisecond accepts millisecond in [0,999]")
	}
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), mt.t.Second(), ms*1_000_000,
		mt.t.Location(),
	)
	return mt
}

// SetNanosecond sets the nanosecond (0–999,999,999). Panics if out of range.
func (mt *MutatingTime) SetNanosecond(nano int) *MutatingTime {
	if nano < 0 || nano > 999_999_999 {
		panic("SetNanosecond accepts nanosecond in [0,999999999]")
	}
	*mt.t = time.Date(
		mt.t.Year(), mt.t.Month(), mt.t.Day(),
		mt.t.Hour(), mt.t.Minute(), mt.t.Second(), nano,
		mt.t.Location(),
	)
	return mt
}

// Time returns the underlying time.Time value.
func (mt *MutatingTime) Time() time.Time {
	return *mt.t
}
