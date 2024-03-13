package years

import "time"

type Clock interface {
	Now() time.Time
}

type stdClock struct{}

func (c *stdClock) Now() time.Time { return time.Now() }
