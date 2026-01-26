package schedule_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years/schedule"
	"github.com/stretchr/testify/assert"
)

func TestSchedule_MatchesDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday should match
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	assert.True(t, s.MatchesDay(mon))

	// Tuesday should match
	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	assert.True(t, s.MatchesDay(tue))

	// Friday should match
	fri := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC)
	assert.True(t, s.MatchesDay(fri))

	// Saturday should not match
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	assert.False(t, s.MatchesDay(sat))

	// Sunday should not match
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	assert.False(t, s.MatchesDay(sun))
}

func TestSchedule_MatchesTime(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// 9:00 should match
	t9 := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	assert.True(t, s.MatchesTime(t9))

	// 12:00 should match
	t12 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	assert.True(t, s.MatchesTime(t12))

	// 16:59 should match
	t16 := time.Date(2024, 1, 15, 16, 59, 0, 0, time.UTC)
	assert.True(t, s.MatchesTime(t16))

	// 8:59 should not match
	t8 := time.Date(2024, 1, 15, 8, 59, 0, 0, time.UTC)
	assert.False(t, s.MatchesTime(t8))

	// 17:00 should not match
	t17 := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)
	assert.False(t, s.MatchesTime(t17))

	// 18:00 should not match
	t18 := time.Date(2024, 1, 15, 18, 0, 0, 0, time.UTC)
	assert.False(t, s.MatchesTime(t18))
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
	assert.True(t, s.MatchesTime(t11))

	// 12:00 should not match (in gap)
	t12 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	assert.False(t, s.MatchesTime(t12))

	// 12:30 should not match (in gap)
	t1230 := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	assert.False(t, s.MatchesTime(t1230))

	// 13:00 should match (after gap)
	t13 := time.Date(2024, 1, 15, 13, 0, 0, 0, time.UTC)
	assert.True(t, s.MatchesTime(t13))
}

func TestSchedule_Contains(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday at 12:00 should be contained
	mon12 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	assert.True(t, s.Contains(mon12))

	// Monday at 8:00 should not be contained (outside hours)
	mon8 := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	assert.False(t, s.Contains(mon8))

	// Saturday at 12:00 should not be contained (not a workday)
	sat12 := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	assert.False(t, s.Contains(sat12))
}

func TestSchedule_PrevMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Sunday -> Friday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	prev := s.PrevMatchingDay(sun)
	assert.Equal(t, time.Friday, prev.Weekday())
	assert.Equal(t, 19, prev.Day())

	// Saturday -> Friday
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	prev = s.PrevMatchingDay(sat)
	assert.Equal(t, time.Friday, prev.Weekday())
	assert.Equal(t, 19, prev.Day())

	// Monday -> Friday (previous week)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	prev = s.PrevMatchingDay(mon)
	assert.Equal(t, time.Friday, prev.Weekday())
	assert.Equal(t, 12, prev.Day())

	// Tuesday -> Monday
	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	prev = s.PrevMatchingDay(tue)
	assert.Equal(t, time.Monday, prev.Weekday())
	assert.Equal(t, 15, prev.Day())
}

func TestSchedule_NextMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Friday -> Monday
	fri := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC)
	next := s.NextMatchingDay(fri)
	assert.Equal(t, time.Monday, next.Weekday())
	assert.Equal(t, 22, next.Day())

	// Saturday -> Monday
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	next = s.NextMatchingDay(sat)
	assert.Equal(t, time.Monday, next.Weekday())
	assert.Equal(t, 22, next.Day())

	// Sunday -> Monday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	next = s.NextMatchingDay(sun)
	assert.Equal(t, time.Monday, next.Weekday())
	assert.Equal(t, 22, next.Day())

	// Monday -> Tuesday
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	next = s.NextMatchingDay(mon)
	assert.Equal(t, time.Tuesday, next.Weekday())
	assert.Equal(t, 16, next.Day())
}

func TestSchedule_PrevNonMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday -> Sunday
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	prev := s.PrevNonMatchingDay(mon)
	assert.Equal(t, time.Sunday, prev.Weekday())
	assert.Equal(t, 14, prev.Day())

	// Tuesday -> Sunday
	tue := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	prev = s.PrevNonMatchingDay(tue)
	assert.Equal(t, time.Sunday, prev.Weekday())
	assert.Equal(t, 14, prev.Day())

	// Sunday -> Saturday
	sun := time.Date(2024, 1, 21, 12, 0, 0, 0, time.UTC)
	prev = s.PrevNonMatchingDay(sun)
	assert.Equal(t, time.Saturday, prev.Weekday())
	assert.Equal(t, 20, prev.Day())
}

func TestSchedule_NextNonMatchingDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Friday -> Saturday
	fri := time.Date(2024, 1, 19, 12, 0, 0, 0, time.UTC)
	next := s.NextNonMatchingDay(fri)
	assert.Equal(t, time.Saturday, next.Weekday())
	assert.Equal(t, 20, next.Day())

	// Saturday -> Sunday
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	next = s.NextNonMatchingDay(sat)
	assert.Equal(t, time.Sunday, next.Weekday())
	assert.Equal(t, 21, next.Day())

	// Monday -> Saturday
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	next = s.NextNonMatchingDay(mon)
	assert.Equal(t, time.Saturday, next.Weekday())
	assert.Equal(t, 20, next.Day())
}

func TestTimeSlot_Duration(t *testing.T) {
	start := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 17, 0, 0, 0, time.UTC)

	slot := schedule.TimeSlot{Start: start, End: end}
	assert.Equal(t, 8*time.Hour, slot.Duration())
}

func TestSchedule_SlotsForDay(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday should return one slot (9-17)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	slots := s.SlotsForDay(mon)
	assert.Len(t, slots, 1)
	assert.Equal(t, 9, slots[0].Start.Hour())
	assert.Equal(t, 17, slots[0].End.Hour())
	assert.Equal(t, 8*time.Hour, slots[0].Duration())

	// Saturday should return no slots
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	slots = s.SlotsForDay(sat)
	assert.Empty(t, slots)
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
	assert.Len(t, slots, 2)

	// First slot: 9-12
	assert.Equal(t, 9, slots[0].Start.Hour())
	assert.Equal(t, 12, slots[0].End.Hour())
	assert.Equal(t, 3*time.Hour, slots[0].Duration())

	// Second slot: 13-17
	assert.Equal(t, 13, slots[1].Start.Hour())
	assert.Equal(t, 17, slots[1].End.Hour())
	assert.Equal(t, 4*time.Hour, slots[1].Duration())
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
	assert.Len(t, slots, 3)

	// First slot: 9-11
	assert.Equal(t, 9, slots[0].Start.Hour())
	assert.Equal(t, 11, slots[0].End.Hour())

	// Second slot: 12-14
	assert.Equal(t, 12, slots[1].Start.Hour())
	assert.Equal(t, 14, slots[1].End.Hour())

	// Third slot: 15-17
	assert.Equal(t, 15, slots[2].Start.Hour())
	assert.Equal(t, 17, slots[2].End.Hour())
}

func TestSchedule_SlotsForRange(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// From Monday to Wednesday (3 workdays)
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)  // Monday
	to := time.Date(2024, 1, 17, 23, 59, 59, 0, time.UTC) // Wednesday

	slots := s.SlotsForRange(from, to)
	assert.Len(t, slots, 3) // 3 workdays = 3 slots

	// Each slot should be 8 hours (9-17)
	for _, slot := range slots {
		assert.Equal(t, 8*time.Hour, slot.Duration())
	}
}

func TestSchedule_SlotsForRange_WithWeekend(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// From Friday to Monday (includes weekend)
	from := time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC)  // Friday
	to := time.Date(2024, 1, 22, 23, 59, 59, 0, time.UTC) // Monday

	slots := s.SlotsForRange(from, to)
	assert.Len(t, slots, 2) // Friday + Monday = 2 slots (weekend excluded)
}

func TestSchedule_AvailableMinutes(t *testing.T) {
	s := schedule.DefaultWorkingHoursSchedule

	// Monday should have 480 minutes (8 hours)
	mon := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	minutes := s.AvailableMinutes(mon)
	assert.Equal(t, 480, minutes)

	// Saturday should have 0 minutes
	sat := time.Date(2024, 1, 20, 12, 0, 0, 0, time.UTC)
	minutes = s.AvailableMinutes(sat)
	assert.Equal(t, 0, minutes)
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
	minutes := s.AvailableMinutes(mon)
	assert.Equal(t, 420, minutes)
}
