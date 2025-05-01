package years_test

import (
	"time"

	"github.com/amberpixels/years"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CoreAliases", func() {
	var base time.Time

	BeforeEach(func() {
		// Use a known Wednesday for predictable week calculations
		base = time.Date(2025, time.May, 7, 15, 30, 45, 123456789, time.UTC)
	})

	DescribeTable("alias functions return expected start times",
		func(alias string, expected time.Time) {
			fn, exists := years.CoreAliases[alias]
			Expect(exists).To(BeTrue(), "alias '%s' should be registered", alias)
			got := fn(base)
			Expect(got).To(Equal(expected), "alias '%s'", alias)
		},

		Entry("today", "today",
			time.Date(2025, time.May, 7, 0, 0, 0, 0, time.UTC)),
		Entry("yesterday", "yesterday",
			time.Date(2025, time.May, 6, 0, 0, 0, 0, time.UTC)),
		Entry("tomorrow", "tomorrow",
			time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)),
		Entry("this-week", "this-week",
			// Sunday of current week
			time.Date(2025, time.May, 4, 0, 0, 0, 0, time.UTC)),
		Entry("last-week", "last-week",
			time.Date(2025, time.April, 27, 0, 0, 0, 0, time.UTC)),
		Entry("next-week", "next-week",
			time.Date(2025, time.May, 11, 0, 0, 0, 0, time.UTC)),
		Entry("next-weekend", "next-weekend",
			time.Date(2025, time.May, 10, 0, 0, 0, 0, time.UTC)),
		Entry("last-weekend", "last-weekend",
			time.Date(2025, time.May, 2, 0, 0, 0, 0, time.UTC)),
		Entry("this-month", "this-month",
			time.Date(2025, time.May, 1, 0, 0, 0, 0, time.UTC)),
		Entry("last-month", "last-month",
			time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC)),
		Entry("next-month", "next-month",
			time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC)),
		Entry("this-year", "this-year",
			time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Entry("last-year", "last-year",
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)),
		Entry("next-year", "next-year",
			time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
})
