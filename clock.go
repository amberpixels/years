package years

import "time"

type Clock interface {
	Now() time.Time
}

type StdClock struct{}

func (c *StdClock) Now() time.Time { return time.Now() }

var stdClock Clock = &StdClock{}

func SetStdClock(c Clock) { stdClock = c }
