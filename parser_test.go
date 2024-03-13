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
	})
})
