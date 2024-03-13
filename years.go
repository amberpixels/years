package years

import (
	"time"
)

// ParseTime calls ParseTime of a default parser
func ParseTime(value string) (time.Time, error) {
	return DefaultParser().ParseTime(value)
}
