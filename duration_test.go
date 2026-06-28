package years_test

import (
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
)

func TestParseISODuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"", 0},
		{"PT45S", 45 * time.Second},
		{"PT15M30S", 15*time.Minute + 30*time.Second},
		{"PT1H2M3S", time.Hour + 2*time.Minute + 3*time.Second},
		{"P1DT2H", 24*time.Hour + 2*time.Hour},
		{"P1DT2H3M4S", 24*time.Hour + 2*time.Hour + 3*time.Minute + 4*time.Second},
	}
	for _, c := range cases {
		got, err := years.ParseISODuration(c.in)
		be.Require(t, err).To(be.Succeed())
		be.Expect(t, got).To(be.Eq(c.want))
	}
}

func TestParseISODuration_Invalid(t *testing.T) {
	for _, in := range []string{"P", "PT", "garbage", "P1W", "P1Y", "1H30M", "PT1H30X"} {
		_, err := years.ParseISODuration(in)
		be.Expect(t, err).To(be.HaveOccurred())
	}
}

func TestFormatDurationClock(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0:00"},
		{30 * time.Second, "0:30"},
		{15*time.Minute + 30*time.Second, "15:30"},
		{time.Hour + 2*time.Minute + 3*time.Second, "1:02:03"},
		{-5 * time.Second, "0:00"},
	}
	for _, c := range cases {
		be.Expect(t, years.FormatDurationClock(c.in)).To(be.Eq(c.want))
	}
}

func TestHumanizeDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{500 * time.Millisecond, "0s"},
		{45 * time.Second, "45s"},
		{90 * time.Minute, "1h 30m"},
		{2*time.Hour + 5*time.Minute + 3*time.Second, "2h 5m"},
		{2*time.Hour + 3*time.Second, "2h 3s"},    // skips zero minutes to next non-zero unit
		{24*time.Hour + 30*time.Minute, "1d 30m"}, // skips zero hours
		{3 * time.Hour, "3h"},                     // single non-zero unit renders alone
		{3*24*time.Hour + 4*time.Hour, "3d 4h"},
		{-45 * time.Second, "45s"}, // magnitude only
	}
	for _, c := range cases {
		be.Expect(t, years.HumanizeDuration(c.in)).To(be.Eq(c.want))
	}
}
