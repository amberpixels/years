package years_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/amberpixels/years"
)

var _ = Describe("ParseLayout", func() {
	DescribeTable("valid Go layouts",
		func(layout string, expectedUnit years.DateUnit, expectedUnits []years.DateUnit) {
			details := years.ParseLayout(layout)
			Expect(details).NotTo(BeNil(), "ParseLayout should not return nil for %s", layout)
			Expect(details.Format).To(Equal(years.LayoutFormatGo))
			Expect(details.MinimalUnit).To(Equal(expectedUnit))
			Expect(details.Units).To(Equal(expectedUnits))
		},

		Entry("full date", "2006-01-02", years.Day, []years.DateUnit{years.Day, years.Month, years.Year}),
		Entry("year-month", "2006-01", years.Month, []years.DateUnit{years.Month, years.Year}),
		Entry("day-month-year", "02-01-2006", years.Day, []years.DateUnit{years.Day, years.Month, years.Year}),
		Entry("year only", "2006", years.Year, []years.DateUnit{years.Year}),
	)

	DescribeTable("Unix timestamp layouts",
		func(layout string, expectedUnit years.DateUnit) {
			details := years.ParseLayout(layout)
			Expect(details).NotTo(BeNil(), "ParseLayout should not return nil for %s", layout)
			Expect(details.Format).To(Equal(years.LayoutFormatUnixTimestamp))
			Expect(details.MinimalUnit).To(Equal(expectedUnit))
			Expect(details.Units).To(Equal([]years.DateUnit{expectedUnit}))
		},

		Entry("seconds", years.LayoutTimestampSeconds, years.UnixSecond),
		Entry("milliseconds", years.LayoutTimestampMilliseconds, years.UnixMillisecond),
		Entry("microseconds", years.LayoutTimestampMicroseconds, years.UnixMicrosecond),
		Entry("nanoseconds", years.LayoutTimestampNanoseconds, years.UnixNanosecond),
	)

	It("returns nil for unknown layout", func() {
		Expect(years.ParseLayout("foo-bar")).To(BeNil())
	})
})
