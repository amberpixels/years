package years_test

import (
	"strconv"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be/be_time"
	. "github.com/onsi/ginkgo/v2" //nolint: revive // ginkgo is fine
	. "github.com/onsi/gomega"    //nolint: revive // gomega is fine
)

var _ = Describe("Parser", func() {
	Context("Default parser", func() {
		AfterEach(func() {
			years.SetParserDefaults(
				years.AcceptUnixSeconds(),
				years.AcceptAliases(),
			)
		})

		It("should parse Unix timestamp", func() {
			var timestamp int64 = 1709682885

			parsedTime, err := years.DefaultParser().JustParse(strconv.Itoa(1709682885))
			Expect(err).Should(Succeed())

			Expect(parsedTime).To(be_time.Unix(timestamp))
		})

		It("should parse DateOnly date", func() {
			timeStr := "2024-03-06"
			expectedTime, _ := time.Parse(time.DateOnly, timeStr)

			years.SetParserDefaults(years.WithLayouts(time.DateOnly))

			parsedTime, err := years.DefaultParser().JustParse(timeStr)
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

			today, err := parser.JustParse("today")
			Expect(err).Should(Succeed())
			Expect(today.String()).To(Equal(`2024-03-01 00:00:00 +0000 UTC`))

			yesterday, err := parser.JustParse("yesterday")
			Expect(err).Should(Succeed())
			Expect(yesterday.String()).To(Equal(`2024-02-29 00:00:00 +0000 UTC`))

			tomorrow, err := parser.JustParse("tomorrow")
			Expect(err).Should(Succeed())
			Expect(tomorrow.String()).To(Equal(`2024-03-02 00:00:00 +0000 UTC`))
		})
	})
})
