package years

import (
	"fmt"
	"time"

	abucast "github.com/amberpixels/abu/cast"
)

// Parse calls Parse of a default parser.
func Parse(layout string, value string) (time.Time, error) {
	return DefaultParser().Parse(layout, value)
}

// JustParse calls JustParse of a default parser.
func JustParse(value string) (time.Time, error) {
	return DefaultParser().JustParse(value)
}

// JustParseRaw attempts to convert or parse any value into a time.Time.
// - If value is time.Time (or custom type convertible to time.Time) the underlined time.Time is returned.
// - If value is a string or custom string type, passes to JustParse.
// - If value implements fmt.Stringer, uses its String() to be parsed via JustParse.
// - If value is a numeric type (int, uint, float), stringifies and passes to JustParse.
// - If value is nil, it returns a zero time.Time.
// Returns an error for unsupported types.
func JustParseRaw(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, nil
	}

	switch {
	case abucast.IsTime(value):
		return abucast.AsTime(value), nil
	case abucast.IsStringish(value):
		return JustParse(abucast.AsString(value))
	case abucast.IsInt(value):
		return JustParse(fmt.Sprint(value))
	}

	return time.Time{}, fmt.Errorf("unsupported type %T for JustParseRaw", value)
}
