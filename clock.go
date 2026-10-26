package years

import "time"

type Clock interface {
	Now() time.Time
}

type StdClock struct{}

func (c *StdClock) Now() time.Time { return time.Now() }

//nolint:gochecknoglobals // package clock, swappable via SetStdClock
var stdClock Clock = &StdClock{}

// SetStdClock sets the package clock. It is meant for tests, to make
// time.Now() mockable.
func SetStdClock(c Clock) { stdClock = c }

// Now returns the current time according to the package-level clock.
// By default it is equivalent to time.Now(); tests can override the source
// via SetStdClock, making any code that relies on years.Now() deterministic.
func Now() time.Time { return stdClock.Now() }
