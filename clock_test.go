package years_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
	"github.com/expectto/be/be_time"
)

func TestStdClock_NowIsCloseToTimeNow(t *testing.T) {
	clock := &years.StdClock{}

	start := time.Now()
	now := clock.Now()
	end := time.Now()

	// The returned time should be between start and end
	be.Expect(t, now).To(be_time.LaterThanEqual(start))
	be.Expect(t, now).To(be_time.EarlierThanEqual(end))
}
