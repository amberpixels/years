package years_test

import (
	"testing"
	"time"

	"github.com/expectto/be"

	"github.com/amberpixels/years"
)

func TestMutatingTime_TruncateToSecond(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
	years.Mutate(&t0).TruncateToSecond()
	be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 30, 13, 45, 59, 0, time.UTC)))
}

func TestMutatingTime_TruncateToMinute(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
	years.Mutate(&t0).TruncateToMinute()
	be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 30, 13, 45, 0, 0, time.UTC)))
}

func TestMutatingTime_TruncateToHour(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
	years.Mutate(&t0).TruncateToHour()
	be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 30, 13, 0, 0, 0, time.UTC)))
}

func TestMutatingTime_TruncateToWeek(t *testing.T) {
	// Wednesday, 2025-04-30.
	t.Run("monday start (ISO)", func(t *testing.T) {
		t0 := time.Date(2025, time.April, 30, 13, 45, 59, 1, time.UTC)
		years.Mutate(&t0).TruncateToWeek(time.Monday)
		be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 28, 0, 0, 0, 0, time.UTC)))
	})

	t.Run("sunday start", func(t *testing.T) {
		t0 := time.Date(2025, time.April, 30, 13, 45, 59, 1, time.UTC)
		years.Mutate(&t0).TruncateToWeek(time.Sunday)
		be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 27, 0, 0, 0, 0, time.UTC)))
	})

	t.Run("on the week-start day stays put (only truncates time)", func(t *testing.T) {
		monday := time.Date(2025, time.April, 28, 9, 30, 0, 0, time.UTC)
		years.Mutate(&monday).TruncateToWeek(time.Monday)
		be.Expect(t, monday).To(be.Eq(time.Date(2025, time.April, 28, 0, 0, 0, 0, time.UTC)))
	})

	t.Run("sunday with monday start wraps to previous monday (ISO)", func(t *testing.T) {
		sunday := time.Date(2025, time.May, 4, 12, 0, 0, 0, time.UTC)
		years.Mutate(&sunday).TruncateToWeek(time.Monday)
		be.Expect(t, sunday).To(be.Eq(time.Date(2025, time.April, 28, 0, 0, 0, 0, time.UTC)))
	})
}

func TestMutatingTime_TruncateToMonth(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 1, time.UTC)
	years.Mutate(&t0).TruncateToMonth()
	be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC)))
}

func TestMutatingTime_TruncateToYear(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 1, time.UTC)
	years.Mutate(&t0).TruncateToYear()
	be.Expect(t, t0).To(be.Eq(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)))
}

func TestMutatingTime_TruncatePreservesLocation(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 1, loc)
	years.Mutate(&t0).TruncateToDay()
	be.Expect(t, t0.Location().String()).To(be.Eq(loc.String()))
	be.Expect(t, t0.Hour()).To(be.Eq(0))
}

// fixedClock is a deterministic Clock for testing years.Now().
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func TestNow_UsesPackageClock(t *testing.T) {
	t.Cleanup(func() { years.SetStdClock(&years.StdClock{}) })

	fixed := time.Date(2030, time.February, 1, 8, 0, 0, 0, time.UTC)
	years.SetStdClock(fixedClock{t: fixed})

	be.Expect(t, years.Now()).To(be.Eq(fixed))
}

func TestMutatingTime_TruncateToDay_SkippedMidnight(t *testing.T) {
	// Santiago enters summer time at midnight on 2026-09-06, so the day starts at 01:00.
	santiago, err := time.LoadLocation("America/Santiago")
	be.Require(t, err).To(be.Succeed())
	want := time.Date(2026, time.September, 6, 1, 0, 0, 0, santiago)

	t0 := time.Date(2026, time.September, 6, 15, 0, 0, 0, santiago)
	years.Mutate(&t0).TruncateToDay()
	be.Expect(t, t0.String()).To(be.Eq(want.String()))

	// Thursday Sep 10; the Sunday-start week begins on the same skipped midnight.
	t1 := time.Date(2026, time.September, 10, 15, 0, 0, 0, santiago)
	years.Mutate(&t1).TruncateToWeek(time.Sunday)
	be.Expect(t, t1.String()).To(be.Eq(want.String()))
}
