package years_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
)

func TestHumanizeFrom(t *testing.T) {
	base := time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		delta time.Duration
		want  string
	}{
		{0, "just now"},
		{-30 * time.Second, "just now"},
		{30 * time.Second, "just now"},
		{-5 * time.Minute, "5m ago"},
		{-3 * time.Hour, "3h ago"},
		{-2 * 24 * time.Hour, "2d ago"},
		{-40 * 24 * time.Hour, "1mo ago"},
		{-400 * 24 * time.Hour, "1y ago"},
		{5 * time.Minute, "in 5m"},
		{3 * time.Hour, "in 3h"},
		{2 * 24 * time.Hour, "in 2d"},
	}
	for _, c := range cases {
		got := years.HumanizeFrom(base, base.Add(c.delta))
		be.Expect(t, got).To(be.Eq(c.want))
	}
}

func TestHumanize_UsesPackageClock(t *testing.T) {
	now := time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC)
	years.SetStdClock(&StaticClock{now: now})
	t.Cleanup(func() { years.SetStdClock(&years.StdClock{}) })

	be.Expect(t, years.Humanize(now.Add(-3*time.Hour))).To(be.Eq("3h ago"))
	be.Expect(t, years.Humanize(now.Add(2*24*time.Hour))).To(be.Eq("in 2d"))
}
