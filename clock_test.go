package years_test

import (
	"testing"
	"time"

	"github.com/expectto/be"
	"github.com/expectto/be/be_time"

	"github.com/amberpixels/years"
)

func TestStdClock_NowIsCloseToTimeNow(t *testing.T) {
	clock := &years.StdClock{}

	start := time.Now()
	now := clock.Now()
	end := time.Now()

	be.Expect(t, now).To(be_time.LaterThanEqual(start))
	be.Expect(t, now).To(be_time.EarlierThanEqual(end))
}
