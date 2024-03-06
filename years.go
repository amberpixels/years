package years

import (
	"time"
)

func ParseTime(value string) (time.Time, error) {
	return DefaultParser().ParseTime(value)
}
