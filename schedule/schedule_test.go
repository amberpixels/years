package schedule_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years/schedule"
	"github.com/expectto/be"
)

func TestSchedule_MatchesDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday should match
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesDay(mon)).To(be.True())

	// Tuesday should match
	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesDay(tue)).To(be.True())

	// Friday should match
	fri := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesDay(fri)).To(be.True())

	// Saturday should not match
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesDay(sat)).To(be.False())

	// Sunday should not match
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesDay(sun)).To(be.False())
}

func TestSchedule_MatchesTime(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// 9:00 should match
	t9 := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t9)).To(be.True())

	// 12:00 should match
	t12 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t12)).To(be.True())

	// 16:59 should match
	t16 := time.Date(2024, 1, 15, 16, 59, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t16)).To(be.True())

	// 8:59 should not match
	t8 := time.Date(2024, 1, 15, 8, 59, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t8)).To(be.False())

	// 17:00 should not match
	t17 := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t17)).To(be.False())

	// 18:00 should not match
	t18 := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t18)).To(be.False())
}

func TestSchedule_MatchesTime_WithGaps(t *testing.T) {
	s := schedule.Schedule{
		Days:      []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartHour: 9,
		EndHour:   17,
		Gaps: []schedule.TimeRange{
			{StartHour: 12, EndHour: 13}, // Lunch break
		},
	}

	// 11:00 should match (before gap)
	t11 := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t11)).To(be.True())

	// 12:00 should not match (in gap)
	t12 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t12)).To(be.False())

	// 12:30 should not match (in gap)
	t1230 := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t1230)).To(be.False())

	// 13:00 should match (after gap)
	t13 := time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC)
	be.Expect(t, s.MatchesTime(t13)).To(be.True())
}

func TestSchedule_Contains(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday at 12:00 should be contained
	mon12 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.Contains(mon12)).To(be.True())

	// Monday at 8:00 should not be contained (outside hours)
	mon8 := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	be.Expect(t, s.Contains(mon8)).To(be.False())

	// Saturday at 12:00 should not be contained (not a workday)
	sat12 := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.Contains(sat12)).To(be.False())
}

func TestSchedule_PrevMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Sunday -> Friday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	prev := s.PrevMatchingDay(sun)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Friday))
	be.Expect(t, prev.Day()).To(be.Eq(19))

	// Saturday -> Friday
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	prev = s.PrevMatchingDay(sat)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Friday))
	be.Expect(t, prev.Day()).To(be.Eq(19))

	// Monday -> Friday (previous week)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	prev = s.PrevMatchingDay(mon)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Friday))
	be.Expect(t, prev.Day()).To(be.Eq(12))

	// Tuesday -> Monday
	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	prev = s.PrevMatchingDay(tue)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Monday))
	be.Expect(t, prev.Day()).To(be.Eq(15))
}

func TestSchedule_NextMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Friday -> Monday
	fri := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC)
	next := s.NextMatchingDay(fri)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Monday))
	be.Expect(t, next.Day()).To(be.Eq(22))

	// Saturday -> Monday
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	next = s.NextMatchingDay(sat)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Monday))
	be.Expect(t, next.Day()).To(be.Eq(22))

	// Sunday -> Monday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	next = s.NextMatchingDay(sun)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Monday))
	be.Expect(t, next.Day()).To(be.Eq(22))

	// Monday -> Tuesday
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	next = s.NextMatchingDay(mon)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Tuesday))
	be.Expect(t, next.Day()).To(be.Eq(16))
}

func TestSchedule_PrevNonMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday -> Sunday
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	prev := s.PrevNonMatchingDay(mon)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Sunday))
	be.Expect(t, prev.Day()).To(be.Eq(14))

	// Tuesday -> Sunday
	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	prev = s.PrevNonMatchingDay(tue)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Sunday))
	be.Expect(t, prev.Day()).To(be.Eq(14))

	// Sunday -> Saturday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	prev = s.PrevNonMatchingDay(sun)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Saturday))
	be.Expect(t, prev.Day()).To(be.Eq(20))
}

func TestSchedule_NextNonMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Friday -> Saturday
	fri := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC)
	next := s.NextNonMatchingDay(fri)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Saturday))
	be.Expect(t, next.Day()).To(be.Eq(20))

	// Saturday -> Sunday
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	next = s.NextNonMatchingDay(sat)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Sunday))
	be.Expect(t, next.Day()).To(be.Eq(21))

	// Monday -> Saturday
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	next = s.NextNonMatchingDay(mon)
	be.Expect(t, next.Weekday()).To(be.Eq(time.Saturday))
	be.Expect(t, next.Day()).To(be.Eq(20))
}

func TestTimeSlot_Duration(t *testing.T) {
	start := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)

	slot := schedule.TimeSlot{Start: start, End: end}
	be.Expect(t, slot.Duration()).To(be.Eq(8 * time.Hour))
}

func TestSchedule_SlotsForDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday should return one slot (9-17)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	slots := s.SlotsForDay(mon)
	be.Require(t, slots).To(be.HaveLength(1))
	be.Expect(t, slots[0].Start.Hour()).To(be.Eq(9))
	be.Expect(t, slots[0].End.Hour()).To(be.Eq(17))
	be.Expect(t, slots[0].Duration()).To(be.Eq(8 * time.Hour))

	// Saturday should return no slots
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	slots = s.SlotsForDay(sat)
	be.Expect(t, slots).To(be.Empty())
}

func TestSchedule_SlotsForDay_WithGaps(t *testing.T) {
	s := schedule.Schedule{
		Days:      []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartHour: 9,
		EndHour:   17,
		Gaps: []schedule.TimeRange{
			{StartHour: 12, EndHour: 13}, // Lunch break
		},
	}

	// Monday should return two slots (9-12 and 13-17)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	slots := s.SlotsForDay(mon)
	be.Require(t, slots).To(be.HaveLength(2))

	// First slot: 9-12
	be.Expect(t, slots[0].Start.Hour()).To(be.Eq(9))
	be.Expect(t, slots[0].End.Hour()).To(be.Eq(12))
	be.Expect(t, slots[0].Duration()).To(be.Eq(3 * time.Hour))

	// Second slot: 13-17
	be.Expect(t, slots[1].Start.Hour()).To(be.Eq(13))
	be.Expect(t, slots[1].End.Hour()).To(be.Eq(17))
	be.Expect(t, slots[1].Duration()).To(be.Eq(4 * time.Hour))
}

func TestSchedule_SlotsForDay_WithMultipleGaps(t *testing.T) {
	s := schedule.Schedule{
		Days:      []time.Weekday{time.Monday},
		StartHour: 9,
		EndHour:   17,
		Gaps: []schedule.TimeRange{
			{StartHour: 11, EndHour: 12}, // Coffee break
			{StartHour: 14, EndHour: 15}, // Afternoon break
		},
	}

	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	slots := s.SlotsForDay(mon)
	be.Require(t, slots).To(be.HaveLength(3))

	// First slot: 9-11
	be.Expect(t, slots[0].Start.Hour()).To(be.Eq(9))
	be.Expect(t, slots[0].End.Hour()).To(be.Eq(11))

	// Second slot: 12-14
	be.Expect(t, slots[1].Start.Hour()).To(be.Eq(12))
	be.Expect(t, slots[1].End.Hour()).To(be.Eq(14))

	// Third slot: 15-17
	be.Expect(t, slots[2].Start.Hour()).To(be.Eq(15))
	be.Expect(t, slots[2].End.Hour()).To(be.Eq(17))
}

func TestSchedule_SlotsForRange(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// From Monday to Wednesday (3 workdays)
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)  // Monday
	to := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC) // Wednesday

	slots := s.SlotsForRange(from, to)
	be.Require(t, slots).To(be.HaveLength(3)) // 3 workdays = 3 slots

	// Each slot should be 8 hours (9-17)
	for _, slot := range slots {
		be.Expect(t, slot.Duration()).To(be.Eq(8 * time.Hour))
	}
}

func TestSchedule_SlotsForRange_WithWeekend(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// From Friday to Monday (includes weekend)
	from := time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)  // Friday
	to := time.Date(2024, 1, 22, 23, 59, 59, 0, time.UTC) // Monday

	slots := s.SlotsForRange(from, to)
	be.Expect(t, slots).To(be.HaveLength(2)) // Friday + Monday = 2 slots (weekend excluded)
}

func TestSchedule_AvailableMinutes(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday should have 480 minutes (8 hours)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.AvailableMinutes(mon)).To(be.Eq(480))

	// Saturday should have 0 minutes
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.AvailableMinutes(sat)).To(be.Eq(0))
}

func TestSchedule_AvailableMinutes_WithGaps(t *testing.T) {
	s := schedule.Schedule{
		Days:      []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartHour: 9,
		EndHour:   17,
		Gaps: []schedule.TimeRange{
			{StartHour: 12, EndHour: 13}, // 1 hour lunch
		},
	}

	// Monday should have 420 minutes (7 hours: 8 - 1 lunch)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, s.AvailableMinutes(mon)).To(be.Eq(420))
}

func TestTimeOfDay_ToDuration(t *testing.T) {
	be.Expect(t, schedule.TimeOfDay{Hour: 6, Minute: 0}.ToDuration()).To(be.Eq(6 * time.Hour))
	be.Expect(t, schedule.TimeOfDay{Hour: 6, Minute: 30}.ToDuration()).To(be.Eq(6*time.Hour + 30*time.Minute))
	be.Expect(t, schedule.TimeOfDay{Hour: 26, Minute: 0}.ToDuration()).To(be.Eq(26 * time.Hour))
}

func TestMultiSlotSchedule_MatchesDay(t *testing.T) {
	ms := schedule.MultiSlotSchedule{
		Days: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
	}

	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	be.Expect(t, ms.MatchesDay(mon)).To(be.True())

	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	be.Expect(t, ms.MatchesDay(tue)).To(be.False())

	wed := time.Date(2024, 1, 17, 12, 0, 0, 0, time.UTC)
	be.Expect(t, ms.MatchesDay(wed)).To(be.True())
}

func TestMultiSlotSchedule_SlotsForDay(t *testing.T) {
	ms := schedule.MultiSlotSchedule{
		Days: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		DaySlots: []schedule.DaySlot{
			{Start: schedule.TimeOfDay{Hour: 6, Minute: 0}, End: schedule.TimeOfDay{Hour: 7, Minute: 30}},
			{Start: schedule.TimeOfDay{Hour: 23, Minute: 0}, End: schedule.TimeOfDay{Hour: 26, Minute: 0}},
		},
	}

	// Monday should return two slots
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	slots := ms.SlotsForDay(mon)
	be.Require(t, slots).To(be.HaveLength(2))

	// First slot: 06:00-07:30 (90 minutes)
	be.Expect(t, slots[0].Start.Hour()).To(be.Eq(6))
	be.Expect(t, slots[0].Start.Minute()).To(be.Eq(0))
	be.Expect(t, slots[0].End.Hour()).To(be.Eq(7))
	be.Expect(t, slots[0].End.Minute()).To(be.Eq(30))
	be.Expect(t, slots[0].Duration()).To(be.Eq(90 * time.Minute))

	// Second slot: 23:00 Mon -> 02:00 Tue (cross-midnight, 3 hours)
	be.Expect(t, slots[1].Start.Hour()).To(be.Eq(23))
	be.Expect(t, slots[1].Start.Minute()).To(be.Eq(0))
	be.Expect(t, slots[1].Start.Day()).To(be.Eq(15)) // Monday
	be.Expect(t, slots[1].End.Hour()).To(be.Eq(2))
	be.Expect(t, slots[1].End.Minute()).To(be.Eq(0))
	be.Expect(t, slots[1].End.Day()).To(be.Eq(16)) // Tuesday (cross-midnight)
	be.Expect(t, slots[1].Duration()).To(be.Eq(3 * time.Hour))

	// Saturday should return no slots
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	slots = ms.SlotsForDay(sat)
	be.Expect(t, slots).To(be.Empty())
}

func TestMultiSlotSchedule_PrevMatchingDay(t *testing.T) {
	ms := schedule.MultiSlotSchedule{
		Days: []time.Weekday{time.Monday, time.Wednesday, time.Friday},
	}

	// Sunday -> Friday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	prev := ms.PrevMatchingDay(sun)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Friday))
	be.Expect(t, prev.Day()).To(be.Eq(19))

	// Thursday -> Wednesday
	thu := time.Date(2024, 1, 18, 12, 0, 0, 0, time.UTC)
	prev = ms.PrevMatchingDay(thu)
	be.Expect(t, prev.Weekday()).To(be.Eq(time.Wednesday))
	be.Expect(t, prev.Day()).To(be.Eq(17))
}

func TestDaySchedule_Interface(t *testing.T) {
	// Verify both types satisfy DaySchedule
	var _ schedule.DaySchedule = schedule.Schedule{}
	var _ schedule.DaySchedule = schedule.MultiSlotSchedule{}
}
