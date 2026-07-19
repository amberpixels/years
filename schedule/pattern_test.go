package schedule_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years/schedule"
	"github.com/expectto/be"
)

// nightSpan is 22:00 through 05:00 next day (cross-midnight).
var nightSpan = schedule.DaySlot{
	Start: schedule.TimeOfDay{Hour: 22, Minute: 0},
	End:   schedule.TimeOfDay{Hour: 29, Minute: 0},
}

// at builds a wall-clock instant on a known week: 2024-01-15 is a Monday.
func at(day int, hour, minute int) time.Time {
	return time.Date(2024, 1, day, hour, minute, 0, 0, time.UTC)
}

func TestWeekPattern_Contains_DaysOnly(t *testing.T) {
	// Days set, Spans empty => the whole day matches
	weekend := schedule.WeekPattern{Days: []time.Weekday{time.Saturday, time.Sunday}}

	be.Expect(t, weekend.Contains(at(20, 0, 0))).To(be.True())    // Saturday midnight
	be.Expect(t, weekend.Contains(at(20, 12, 30))).To(be.True())  // Saturday noon
	be.Expect(t, weekend.Contains(at(21, 23, 59))).To(be.True())  // Sunday last minute
	be.Expect(t, weekend.Contains(at(19, 23, 59))).To(be.False()) // Friday last minute
	be.Expect(t, weekend.Contains(at(22, 0, 0))).To(be.False())   // Monday midnight
}

func TestWeekPattern_Contains_SpansOnly_MidnightBoundaries(t *testing.T) {
	// Spans set, Days empty => every day; cross-midnight span behaves as pure
	// minute-of-day wraparound.
	nights := schedule.WeekPattern{Spans: []schedule.DaySlot{nightSpan}}

	be.Expect(t, nights.Contains(at(15, 21, 59))).To(be.False()) // just before
	be.Expect(t, nights.Contains(at(15, 22, 0))).To(be.True())   // inclusive start
	be.Expect(t, nights.Contains(at(15, 23, 10))).To(be.True())  // before midnight
	be.Expect(t, nights.Contains(at(16, 0, 30))).To(be.True())   // after midnight
	be.Expect(t, nights.Contains(at(16, 4, 59))).To(be.True())   // last minute
	be.Expect(t, nights.Contains(at(16, 5, 0))).To(be.False())   // exclusive end
	be.Expect(t, nights.Contains(at(16, 12, 0))).To(be.False())  // midday
}

func TestWeekPattern_Contains_AttributionRule(t *testing.T) {
	// A cross-midnight span belongs to the day it starts on: "sunday nights"
	// includes Monday 01:00 but not Sunday 01:00 (that night started Saturday).
	sundayNights := schedule.WeekPattern{
		Days:  []time.Weekday{time.Sunday},
		Spans: []schedule.DaySlot{nightSpan},
	}

	be.Expect(t, sundayNights.Contains(at(21, 22, 30))).To(be.True())  // Sunday 22:30
	be.Expect(t, sundayNights.Contains(at(22, 1, 0))).To(be.True())    // Monday 01:00 (attributed to Sunday)
	be.Expect(t, sundayNights.Contains(at(22, 4, 59))).To(be.True())   // Monday 04:59
	be.Expect(t, sundayNights.Contains(at(21, 1, 0))).To(be.False())   // Sunday 01:00 (Saturday's night)
	be.Expect(t, sundayNights.Contains(at(22, 5, 0))).To(be.False())   // Monday 05:00 (night over)
	be.Expect(t, sundayNights.Contains(at(21, 12, 0))).To(be.False())  // Sunday noon (not a night)
	be.Expect(t, sundayNights.Contains(at(22, 22, 30))).To(be.False()) // Monday 22:30 (Monday's night)
}

func TestWeekPattern_Contains_WeekWrap(t *testing.T) {
	// Saturday nights wrap the week boundary into Sunday morning.
	saturdayNights := schedule.WeekPattern{
		Days:  []time.Weekday{time.Saturday},
		Spans: []schedule.DaySlot{nightSpan},
	}

	be.Expect(t, saturdayNights.Contains(at(20, 23, 30))).To(be.True()) // Sat 23:30
	be.Expect(t, saturdayNights.Contains(at(21, 0, 30))).To(be.True())  // Sun 00:30 (Saturday's night)
	be.Expect(t, saturdayNights.Contains(at(21, 23, 30))).To(be.False())
}

func TestWeekPattern_Contains_Empty(t *testing.T) {
	// Zero pattern: every day, whole day.
	all := schedule.WeekPattern{}

	be.Expect(t, all.Contains(at(15, 0, 0))).To(be.True())
	be.Expect(t, all.Contains(at(21, 23, 59))).To(be.True())
}

func TestWeekPatterns_Contains_Union(t *testing.T) {
	// "sundays,mornings": sundays all day OR mornings every day
	union := schedule.WeekPatterns{
		{Days: []time.Weekday{time.Sunday}},
		{Spans: []schedule.DaySlot{{
			Start: schedule.TimeOfDay{Hour: 5, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 12, Minute: 0},
		}}},
	}

	be.Expect(t, union.Contains(at(21, 18, 0))).To(be.True())  // Sunday evening
	be.Expect(t, union.Contains(at(16, 9, 0))).To(be.True())   // Tuesday morning
	be.Expect(t, union.Contains(at(16, 18, 0))).To(be.False()) // Tuesday evening

	be.Expect(t, schedule.WeekPatterns{}.Contains(at(15, 12, 0))).To(be.False())
}

func TestWeekPattern_Validate(t *testing.T) {
	valid := schedule.WeekPattern{
		Days:  []time.Weekday{time.Sunday},
		Spans: []schedule.DaySlot{nightSpan},
	}
	be.Expect(t, valid.Validate()).To(be.Succeed())

	be.Expect(t, schedule.WeekPattern{}.Validate()).To(be.Succeed())

	badDay := schedule.WeekPattern{Days: []time.Weekday{time.Weekday(7)}}
	be.Expect(t, badDay.Validate()).To(be.HaveOccurred())

	badMinute := schedule.WeekPattern{Spans: []schedule.DaySlot{{
		Start: schedule.TimeOfDay{Hour: 9, Minute: 60},
		End:   schedule.TimeOfDay{Hour: 17, Minute: 0},
	}}}
	be.Expect(t, badMinute.Validate()).To(be.HaveOccurred())

	startsNextDay := schedule.WeekPattern{Spans: []schedule.DaySlot{{
		Start: schedule.TimeOfDay{Hour: 24, Minute: 0},
		End:   schedule.TimeOfDay{Hour: 26, Minute: 0},
	}}}
	be.Expect(t, startsNextDay.Validate()).To(be.HaveOccurred())

	endBeforeStart := schedule.WeekPattern{Spans: []schedule.DaySlot{{
		Start: schedule.TimeOfDay{Hour: 17, Minute: 0},
		End:   schedule.TimeOfDay{Hour: 9, Minute: 0},
	}}}
	be.Expect(t, endBeforeStart.Validate()).To(be.HaveOccurred())

	tooLong := schedule.WeekPattern{Spans: []schedule.DaySlot{{
		Start: schedule.TimeOfDay{Hour: 1, Minute: 0},
		End:   schedule.TimeOfDay{Hour: 48, Minute: 30},
	}}}
	be.Expect(t, tooLong.Validate()).To(be.HaveOccurred())
}

func TestWeekPatterns_Validate(t *testing.T) {
	ok := schedule.WeekPatterns{{Days: []time.Weekday{time.Monday}}, {}}
	be.Expect(t, ok.Validate()).To(be.Succeed())

	bad := schedule.WeekPatterns{{}, {Days: []time.Weekday{time.Weekday(-1)}}}
	be.Expect(t, bad.Validate()).To(be.HaveOccurred())
}

func TestWeekPattern_MultiSlotSchedule(t *testing.T) {
	pattern := schedule.WeekPattern{
		Days:  []time.Weekday{time.Monday},
		Spans: []schedule.DaySlot{nightSpan},
	}
	ms := pattern.MultiSlotSchedule(time.UTC)

	mon := at(15, 12, 0)
	slots := ms.SlotsForDay(mon)
	be.Require(t, slots).To(be.HaveLength(1))
	be.Expect(t, slots[0].Start.Hour()).To(be.Eq(22))
	be.Expect(t, slots[0].End.Day()).To(be.Eq(16)) // crosses into Tuesday
	be.Expect(t, slots[0].End.Hour()).To(be.Eq(5))

	// Empty Days/Spans expand into every day / the whole day.
	full := schedule.WeekPattern{}.MultiSlotSchedule(time.UTC)
	for day := 15; day < 22; day++ {
		daySlots := full.SlotsForDay(at(day, 0, 0))
		be.Require(t, daySlots).To(be.HaveLength(1))
		be.Expect(t, daySlots[0].Duration()).To(be.Eq(24 * time.Hour))
	}
}
