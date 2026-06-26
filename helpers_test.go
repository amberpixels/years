package years_test

import (
	"time"
)

const (
	TestDataPath = "internal/testdata"
)

type StaticClock struct {
	now time.Time
}

func (c *StaticClock) Now() time.Time { return c.now }
