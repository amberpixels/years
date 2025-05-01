package years_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be/be_time"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JustParseRaw", func() {
	Context("when value is nil", func() {
		It("returns zero time and no error", func() {
			parsedTime, err := years.JustParseRaw(nil)
			Expect(err).Should(Succeed())
			Expect(parsedTime).To(Equal(time.Time{}))
		})
	})

	Context("when value is a time.Time", func() {
		It("returns the same time value", func() {
			now := time.Now().Truncate(time.Second)
			parsedTime, err := years.JustParseRaw(now)
			Expect(err).Should(Succeed())
			Expect(parsedTime).To(Equal(now))
		})
	})

	Context("when value is a numeric string", func() {
		It("parses it as a Unix timestamp", func() {
			var timestamp int64 = 1709682885
			value := strconv.Itoa(int(timestamp))
			parsedTime, err := years.JustParseRaw(value)
			Expect(err).Should(Succeed())
			Expect(parsedTime).To(be_time.Unix(timestamp))
		})
	})

	Context("when value is an integer type", func() {
		It("parses the integer as a Unix timestamp", func() {
			var v int = 12345
			parsedTime, err := years.JustParseRaw(v)
			Expect(err).Should(Succeed())
			Expect(parsedTime).To(be_time.Unix(int64(v)))
		})
	})

	Context("when value is an unsupported type", func() {
		It("returns an unsupported type error", func() {
			unsupported := 3.14
			_, err := years.JustParseRaw(unsupported)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				fmt.Sprintf("unsupported type %T for JustParseRaw", unsupported),
			))
		})
	})
})
