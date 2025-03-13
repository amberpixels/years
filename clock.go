package years

import "time"

type Clock interface {
	Now() time.Time
}

type StdClock struct{}

func (c *StdClock) Now() time.Time { return time.Now() }

//nolint:gochecknoglobals // it's ok
var stdClock Clock = &StdClock{}

// SetStdClock sets the default clock to use.
// Note: this considered to be called from tests, so time.Now() is mockable.
func SetStdClock(c Clock) { stdClock = c }
