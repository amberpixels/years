package years_test

import (
	"fmt"
	"github.com/amberpixels/years"
	"github.com/expectto/be/be_time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Parser", func() {
	Context("Default parser", func() {
		It("should parse Unix timestamp", func() {
			var timestamp int64 = 1709682885

			parsedTime, err := years.DefaultParser().ParseTime(fmt.Sprintf("%d", 1709682885))
			Expect(err).Should(Succeed())

			Expect(parsedTime).To(be_time.Unix(timestamp))
		})

		It("should parse DateOnly date", func() {
			timeStr := "2024-03-06"
			expectedTime, _ := time.Parse(time.DateOnly, timeStr)

			years.SetDefaults(years.WithLayouts(time.DateOnly))

			parsedTime, err := years.DefaultParser().ParseTime(timeStr)
			Expect(err).Should(Succeed())
			Expect(parsedTime).To(Equal(expectedTime))
		})

		It("should parse today/yesterday/tomorrow alias", func() {
			mockClock := &StaticClock{
				now: time.Date(2024, time.March, 01, 14, 30, 59, 0, time.UTC),
			}
			parser := years.NewParser(
				years.WithCustomClock(mockClock),
				years.AcceptAliases(),
				years.AcceptUnixSeconds(),
			)

			today, err := parser.ParseTime("today")
			Expect(err).Should(Succeed())
			Expect(today.String()).To(Equal(`2024-03-01 00:00:00 +0000 UTC`))

			yesterday, err := parser.ParseTime("yesterday")
			Expect(err).Should(Succeed())
			Expect(yesterday.String()).To(Equal(`2024-02-29 00:00:00 +0000 UTC`))

			tomorrow, err := parser.ParseTime("tomorrow")
			Expect(err).Should(Succeed())
			Expect(tomorrow.String()).To(Equal(`2024-03-02 00:00:00 +0000 UTC`))
		})
	})
})
