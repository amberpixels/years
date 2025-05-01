package years_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/amberpixels/years"
)

var _ = Describe("MutatingTime", func() {
	var t0 time.Time

	BeforeEach(func() {
		t0 = time.Date(2025, time.April, 30, 13, 45, 59, 123456789, time.UTC)
	})

	Describe("TruncateToDay", func() {
		It("zeros hour, minute, second, and nanosecond", func() {
			years.Mutate(&t0).TruncateToDay()
			Expect(t0).To(Equal(time.Date(2025, time.April, 30, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("SetYear", func() {
		It("updates only the year", func() {
			years.Mutate(&t0).SetYear(2021)
			Expect(t0.Year()).To(Equal(2021))
			Expect(t0.Month()).To(Equal(time.April))
			Expect(t0.Day()).To(Equal(30))
			Expect(t0.Hour()).To(Equal(13))
		})
	})

	Describe("SetMonth", func() {
		It("updates only the month", func() {
			t0 = time.Date(2021, time.January, 15, 0, 0, 0, 0, time.UTC)
			years.Mutate(&t0).SetMonth(time.December)
			Expect(t0).To(Equal(time.Date(2021, time.December, 15, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("SetDay", func() {
		Context("normal day change", func() {
			BeforeEach(func() {
				t0 = time.Date(2021, time.March, 10, 5, 6, 7, 8, time.UTC)
			})
			It("updates only the day", func() {
				years.Mutate(&t0).SetDay(20)
				Expect(t0).To(Equal(time.Date(2021, time.March, 20, 5, 6, 7, 8, time.UTC)))
			})
		})

		Context("overflow past month boundary", func() {
			BeforeEach(func() {
				t0 = time.Date(2021, time.February, 28, 0, 0, 0, 0, time.UTC)
			})
			It("normalizes correctly", func() {
				years.Mutate(&t0).SetDay(31)
				Expect(t0).To(Equal(time.Date(2021, time.March, 3, 0, 0, 0, 0, time.UTC)))
			})
		})
	})

	Describe("SetHour, SetMinute, SetSecond, SetNanosecond", func() {
		It("sets all units correctly", func() {
			t0 = time.Date(2025, time.July, 4, 0, 0, 0, 0, time.UTC)
			years.Mutate(&t0).
				SetHour(23).
				SetMinute(59).
				SetSecond(58).
				SetNanosecond(123456789)
			Expect(t0).To(Equal(time.Date(2025, time.July, 4, 23, 59, 58, 123456789, time.UTC)))
		})

		It("panics on invalid hour", func() {
			Expect(func() { years.Mutate(&t0).SetHour(-1) }).To(Panic())
			Expect(func() { years.Mutate(&t0).SetHour(24) }).To(Panic())
		})

		It("panics on invalid minute", func() {
			Expect(func() { years.Mutate(&t0).SetMinute(-1) }).To(Panic())
			Expect(func() { years.Mutate(&t0).SetMinute(60) }).To(Panic())
		})

		It("panics on invalid second", func() {
			Expect(func() { years.Mutate(&t0).SetSecond(-1) }).To(Panic())
			Expect(func() { years.Mutate(&t0).SetSecond(60) }).To(Panic())
		})

		It("panics on invalid nanosecond", func() {
			Expect(func() { years.Mutate(&t0).SetNanosecond(-1) }).To(Panic())
			Expect(func() { years.Mutate(&t0).SetNanosecond(1_000_000_000) }).To(Panic())
		})
	})

	Describe("Time method", func() {
		It("returns the underlying time value", func() {
			original := time.Date(1999, time.December, 31, 23, 59, 59, 999, time.UTC)
			mt := years.Mutate(&original)
			Expect(mt.Time()).To(Equal(original))
		})
	})

	Describe("Fluent chaining", func() {
		It("applies all mutations in a row", func() {
			t0 = time.Date(2000, time.January, 1, 1, 2, 3, 4, time.UTC)
			years.Mutate(&t0).
				SetYear(2021).
				SetMonth(time.December).
				SetDay(25).
				SetHour(10).
				SetMinute(20).
				SetSecond(30).
				SetMillisecond(400)

			Expect(t0).To(Equal(time.Date(2021, time.December, 25, 10, 20, 30, 400_000_000, time.UTC)))
		})
	})
})
