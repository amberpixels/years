package years_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/amberpixels/years"
)

var _ = Describe("StdClock", func() {
	var clock *years.StdClock

	BeforeEach(func() {
		clock = &years.StdClock{}
	})

	It("returns current time close to time.Now()", func() {
		start := time.Now()
		now := clock.Now()
		end := time.Now()

		// The returned time should be between start and end
		Expect(now).To(BeTemporally(">=", start))
		Expect(now).To(BeTemporally("<=", end))
	})
})
