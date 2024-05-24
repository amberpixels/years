package years

import "time"

// MutatingTime is a wrapper of a standard time.Time so it can be mutated via helper methods
//
// Example:
//
//	t, _ := time.Parse("...")
//	Mutate(&t).SomeModifyingMethod() // leads to update the t
type MutatingTime struct {
	t *time.Time
}

func Mutate(v *time.Time) *MutatingTime {
	return &MutatingTime{v}
}

// TruncateToDay overrides hour, minute, second, nanosecond to zero
func (t *MutatingTime) TruncateToDay() *MutatingTime {
	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), 0, 0, 0, 0, t.t.Location())

	return t
}

// SetYear overrides year of the time
func (t *MutatingTime) SetYear(v int) *MutatingTime {
	*(t.t) = time.Date(0, t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		AddDate(v, 0, 0)

	return t
}

// SetMonth overrides month of the time
func (t *MutatingTime) SetMonth(month time.Month) *MutatingTime {
	*(t.t) = time.Date(t.t.Year(), 0, t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		AddDate(0, int(month), 0)

	return t
}

// SetDay overrides day of the time
// Note: Feb2 .SetDay(31) will lead to ~Mar2-3 (depending on days in Feb)
func (t *MutatingTime) SetDay(day int) *MutatingTime {
	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		AddDate(0, 0, day)

	return t
}

// SetHour overrides hour of the time
func (t *MutatingTime) SetHour(hour int) *MutatingTime {
	if hour < 0 || hour >= 24 {
		panic("SetMinute accepts hour to be from 0 to 23")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), 0, t.t.Minute(), t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		Add(time.Duration(hour) * time.Hour)

	return t
}

// SetMinute overrides minute of the time
func (t *MutatingTime) SetMinute(minute int) *MutatingTime {
	if minute < 0 || minute >= 60 {
		panic("SetMinute accepts minute to be from 0 to 59")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), 0, t.t.Second(), t.t.Nanosecond(), t.t.Location()).
		Add(time.Duration(minute) * time.Minute)

	return t
}

// SetSecond overrides second of the time
func (t *MutatingTime) SetSecond(second int) *MutatingTime {
	if second < 0 || second >= 60 {
		panic("SetMinute accepts second to be from 0 to 59")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), 0, t.t.Nanosecond(), t.t.Location()).
		Add(time.Duration(second) * time.Second)

	return t
}

// SetNanosecond overrides nanosecond of the time
func (t *MutatingTime) SetNanosecond(nanosecond int) *MutatingTime {
	if nanosecond < 0 || nanosecond >= 60 {
		panic("SetMinute accepts nanosecond to be from 0 to 59")
	}

	*(t.t) = time.Date(t.t.Year(), t.t.Month(), t.t.Day(), t.t.Hour(), t.t.Minute(), t.t.Second(), 0, t.t.Location()).
		Add(time.Duration(nanosecond))

	return t
}

func (t *MutatingTime) Time() time.Time { return *(t.t) }
