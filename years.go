package years

import (
	"time"
)

// Parse calls Parse of a default parser.
func Parse(layout string, value string) (time.Time, error) {
	return DefaultParser().Parse(layout, value)
}

// JustParse calls JustParse of a default parser.
func JustParse(value string) (time.Time, error) {
	return DefaultParser().JustParse(value)
}
