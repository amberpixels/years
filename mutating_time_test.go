package years_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
)

func TestMutatingTime_TruncateToDay(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
	years.Mutate(&t0).TruncateToDay()
	be.Expect(t, t0).To(be.Eq(time.Date(2025, time.April, 30, 0, 0, 0, 0, time.UTC)))
}

func TestMutatingTime_SetYear(t *testing.T) {
	t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
	years.Mutate(&t0).SetYear(2021)
	be.Expect(t, t0.Year()).To(be.Eq(2021))
	be.Expect(t, t0.Month()).To(be.Eq(time.April))
	be.Expect(t, t0.Day()).To(be.Eq(30))
	be.Expect(t, t0.Hour()).To(be.Eq(13))
}

func TestMutatingTime_SetMonth(t *testing.T) {
	t0 := time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC)
	years.Mutate(&t0).SetMonth(time.December)
	be.Expect(t, t0).To(be.Eq(time.Date(2021, time.December, 15, 0, 0, 0, 0, time.UTC)))
}

func TestMutatingTime_SetDay(t *testing.T) {
	t.Run("normal day change", func(t *testing.T) {
		t0 := time.Date(2021, time.March, 10, 5, 6, 7, 8, time.UTC)
		years.Mutate(&t0).SetDay(20)
		be.Expect(t, t0).To(be.Eq(time.Date(2021, time.March, 20, 5, 6, 7, 8, time.UTC)))
	})

	t.Run("overflow past month boundary", func(t *testing.T) {
		t0 := time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC)
		years.Mutate(&t0).SetDay(31)
		be.Expect(t, t0).To(be.Eq(time.Date(2021, time.March, 3, 0, 0, 0, 0, time.UTC)))
	})
}

func TestMutatingTime_SetTimeUnits(t *testing.T) {
	t.Run("sets all units correctly", func(t *testing.T) {
		t0 := time.Date(2025, time.July, 4, 0, 0, 0, 0, time.UTC)
		years.Mutate(&t0).
			SetHour(23).
			SetMinute(59).
			SetSecond(58).
			SetNanosecond(123456789)
		be.Expect(t, t0).To(be.Eq(time.Date(2025, time.July, 4, 23, 59, 58, 123456789, time.UTC)))
	})

	t.Run("panics on invalid hour", func(t *testing.T) {
		t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
		be.Expect(t, func() { years.Mutate(&t0).SetHour(-1) }).To(be.Panic())
		be.Expect(t, func() { years.Mutate(&t0).SetHour(24) }).To(be.Panic())
	})

	t.Run("panics on invalid minute", func(t *testing.T) {
		t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
		be.Expect(t, func() { years.Mutate(&t0).SetMinute(-1) }).To(be.Panic())
		be.Expect(t, func() { years.Mutate(&t0).SetMinute(60) }).To(be.Panic())
	})

	t.Run("panics on invalid second", func(t *testing.T) {
		t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
		be.Expect(t, func() { years.Mutate(&t0).SetSecond(-1) }).To(be.Panic())
		be.Expect(t, func() { years.Mutate(&t0).SetSecond(60) }).To(be.Panic())
	})

	t.Run("panics on invalid nanosecond", func(t *testing.T) {
		t0 := time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
		be.Expect(t, func() { years.Mutate(&t0).SetNanosecond(-1) }).To(be.Panic())
		be.Expect(t, func() { years.Mutate(&t0).SetNanosecond(1_000_000_000) }).To(be.Panic())
	})
}

func TestMutatingTime_Time(t *testing.T) {
	original := time.Date(1999, time.December, 31, 23, 59, 59, 999, time.UTC)
	mt := years.Mutate(&original)
	be.Expect(t, mt.Time()).To(be.Eq(original))
}

func TestMutatingTime_FluentChaining(t *testing.T) {
	t0 := time.Date(2000, time.January, 1, 1, 2, 3, 4, time.UTC)
	years.Mutate(&t0).
		SetYear(2021).
		SetMonth(time.December).
		SetDay(25).
		SetHour(10).
		SetMinute(20).
		SetSecond(30).
		SetMillisecond(400)

	be.Expect(t, t0).To(be.Eq(time.Date(2021, time.December, 25, 10, 20, 30, 400_000_000, time.UTC)))
}
