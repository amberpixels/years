package schedule_test

import (
	"testing"
	"time"

	"github.com/expectto/be"
	"github.com/expectto/be/be_string"

	"github.com/amberpixels/years/schedule"
)

func TestDayPartition_Validate_Default(t *testing.T) {
	be.Expect(t, schedule.DefaultDayPartition.Validate()).To(be.Succeed())
}

func TestDayPartition_Validate_Empty(t *testing.T) {
	be.Expect(t, schedule.DayPartition{}.Validate()).To(be.HaveOccurred())
}

func TestDayPartition_Validate_Gap(t *testing.T) {
	// 23:00-24:00 is covered by nobody
	gappy := schedule.DayPartition{
		{Name: "day", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 0, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 12, Minute: 0},
		}},
		{Name: "rest", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 12, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 23, Minute: 0},
		}},
	}

	err := gappy.Validate()
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring("no part covers 23:00"))
}

func TestDayPartition_Validate_Overlap(t *testing.T) {
	overlapping := schedule.DayPartition{
		{Name: "am", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 0, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 13, Minute: 0},
		}},
		{Name: "pm", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 12, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 24, Minute: 0},
		}},
	}

	err := overlapping.Validate()
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring(`"am" and "pm" overlap at 12:00`))
}

func TestDayPartition_Validate_WrappedOverlap(t *testing.T) {
	// The wrapping night (22:00-05:00) collides with a morning starting at 04:00.
	wrapped := schedule.DayPartition{
		{Name: "night", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 22, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 29, Minute: 0},
		}},
		{Name: "morning", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 4, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 22, Minute: 0},
		}},
	}

	err := wrapped.Validate()
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring(`"night" and "morning" overlap at 04:00`))
}

func TestDayPartition_Validate_BadSpan(t *testing.T) {
	bad := schedule.DayPartition{
		{Name: "broken", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 12, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 12, Minute: 0}, // zero-length
		}},
	}

	err := bad.Validate()
	be.Require(t, err).To(be.HaveOccurred())
	be.Expect(t, err.Error()).To(be_string.ContainingSubstring(`part "broken"`))

	unnamed := schedule.DayPartition{
		{Name: "", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 0, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 24, Minute: 0},
		}},
	}
	be.Expect(t, unnamed.Validate()).To(be.HaveOccurred())
}

func TestDayPartition_PartOf(t *testing.T) {
	dp := schedule.DefaultDayPartition

	cases := map[string]struct {
		hour, minute int
		want         string
	}{
		"pre-dawn wraps into night": {2, 30, "night"},
		"night last minute":         {4, 59, "night"},
		"morning starts":            {5, 0, "morning"},
		"late morning":              {11, 59, "morning"},
		"lunch":                     {12, 0, "lunchtime"},
		"afternoon":                 {14, 0, "afternoon"},
		"evening":                   {17, 0, "evening"},
		"night starts":              {22, 0, "night"},
		"before midnight":           {23, 59, "night"},
	}
	for name, tc := range cases {
		part := dp.PartOf(at(15, tc.hour, tc.minute))
		be.Expect(t, part.Name).To(be.Eq(tc.want), name)
	}
}

func TestDayPartition_PartOf_Totality(t *testing.T) {
	// The Validate invariant in action: every minute of a day belongs to
	// exactly one part - that's what makes PartOf a safe group-by key.
	dp := schedule.DefaultDayPartition
	be.Require(t, dp.Validate()).To(be.Succeed())

	counts := make(map[string]int)
	for m := range 1440 {
		part := dp.PartOf(at(15, m/60, m%60))
		be.Require(t, part.Name).To(be_string.NonEmptyString())
		counts[part.Name]++
	}

	// Spot-check the per-part minute budgets sum up to the full day.
	be.Expect(t, counts["night"]).To(be.Eq(7 * 60))
	be.Expect(t, counts["morning"]).To(be.Eq(7 * 60))
	be.Expect(t, counts["lunchtime"]).To(be.Eq(2 * 60))
	be.Expect(t, counts["afternoon"]).To(be.Eq(3 * 60))
	be.Expect(t, counts["evening"]).To(be.Eq(5 * 60))
}

func TestDayPartition_PartOf_InvalidPartitionZeroValue(t *testing.T) {
	// A partition with a gap (which Validate would reject) returns the zero
	// DayPart for uncovered minutes instead of misclassifying.
	gappy := schedule.DayPartition{
		{Name: "day", Span: schedule.DaySlot{
			Start: schedule.TimeOfDay{Hour: 0, Minute: 0},
			End:   schedule.TimeOfDay{Hour: 12, Minute: 0},
		}},
	}

	be.Expect(t, gappy.PartOf(at(15, 15, 0))).To(be.Eq(schedule.DayPart{}))
}

func TestDayPartition_PartOf_IgnoresLocationConversion(t *testing.T) {
	// PartOf reads the wall clock of t as-is - no timezone conversion.
	cet := time.FixedZone("CET", 3600)

	morningInCET := time.Date(2024, 1, 15, 9, 0, 0, 0, cet)
	be.Expect(t, schedule.DefaultDayPartition.PartOf(morningInCET).Name).To(be.Eq("morning"))
}
