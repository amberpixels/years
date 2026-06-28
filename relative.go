package years

import (
	"fmt"
	"time"
)

// Coarse relative-time buckets. These are approximations for friendly display,
// not precise calendar math (a "month" is 30 days, a "year" is 365).
const (
	approxMonth = 30 * 24 * time.Hour
	approxYear  = 365 * 24 * time.Hour
)

// Humanize renders t relative to the current time (the package clock, see Now),
// automatically choosing past ("3d ago") or future ("in 3d") phrasing; gaps under
// a minute render as "just now". It is the Go counterpart to the "time ago"
// helpers usually hand-written on the frontend. For deterministic results in
// tests, override the clock via SetStdClock or use HumanizeFrom.
func Humanize(t time.Time) string { return HumanizeFrom(Now(), t) }

// HumanizeFrom is Humanize with an explicit base time instead of the package
// clock, so it needs no clock and is convenient for tests: it describes t as seen
// from base ("3d ago" when t precedes base, "in 3d" when it follows).
func HumanizeFrom(base, t time.Time) string {
	d := t.Sub(base)
	if d < 0 {
		if -d < time.Minute {
			return "just now"
		}
		return relativeMagnitude(-d) + " ago"
	}
	if d < time.Minute {
		return "just now"
	}
	return "in " + relativeMagnitude(d)
}

// relativeMagnitude renders a non-negative duration as a single coarse unit
// (m/h/d/mo/y) for relative-time phrasing. Callers add the "ago"/"in" framing.
func relativeMagnitude(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d/time.Minute))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d/time.Hour))
	case d < approxMonth:
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	case d < approxYear:
		return fmt.Sprintf("%dmo", int(d/approxMonth))
	default:
		return fmt.Sprintf("%dy", int(d/approxYear))
	}
}
