package years

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

// DateUnit is the unit of a date (Day, Month, Year) or of a unix timestamp.
type DateUnit int

const UnitUndefined DateUnit = 0

const (
	// Day as day of the month
	// TODO(nice-to-have) support day of the week + day of the year.
	Day  DateUnit = 1 << iota
	Week          // not supported yet
	Month
	Quarter // not supported yet
	Year

	// UnixSecond as well as UnixMillisecond, UnixMicrosecond, UnixNanosecond
	// are special units for Unix timestamps.
	UnixSecond
	UnixMillisecond
	UnixMicrosecond
	UnixNanosecond
)

func (du DateUnit) String() string {
	switch du {
	case UnitUndefined:
		return ""

	case Day:
		return "day"
	case Month:
		return "month"
	case Week, Quarter:
		panic("not-implemented") // todo
	case Year:
		return "year"
	case UnixSecond:
		return "unix_second"
	case UnixMillisecond:
		return "unix_millisecond"
	case UnixMicrosecond:
		return "unix_microsecond"
	case UnixNanosecond:
		return "unix_nanosecond"
	default:
		panic("fix DateUnit enum!")
	}
}

func (du DateUnit) Defined() bool { return du != UnitUndefined }

// DateUnitsDict holds all available DateUnits.
//
//nolint:gochecknoglobals // read-only enum dictionary
var DateUnitsDict = struct {
	Day   DateUnit
	Month DateUnit
	Year  DateUnit

	UnixSecond      DateUnit
	UnixMillisecond DateUnit
	UnixMicrosecond DateUnit
	UnixNanosecond  DateUnit
}{
	Day:   Day,
	Month: Month,
	Year:  Year,

	// TODO: support week and quarter

	UnixSecond:      UnixSecond,
	UnixMillisecond: UnixMillisecond,
	UnixMicrosecond: UnixMicrosecond,
	UnixNanosecond:  UnixNanosecond,
}

type LayoutFormat int

const LayoutFormatUndefined LayoutFormat = 0

const (
	// LayoutFormatGo is a format that is supported by Go time.Parse.
	LayoutFormatGo LayoutFormat = 1 << iota
	// LayoutFormatUnixTimestamp is a format that parses time from a Unix timestamp (seconds to nanoseconds).
	LayoutFormatUnixTimestamp

	// TODO(nice-to-have): support more formats, e.g. JS-like formats (YYYY, etc).
)

func (lf LayoutFormat) String() string {
	switch lf {
	case LayoutFormatGo:
		return "go"
	case LayoutFormatUnixTimestamp:
		return "unix_timestamp"
	case LayoutFormatUndefined:
		panic("fix LayoutFormat enum!")
	default:
		panic("fix LayoutFormat enum!")
	}
}

// LayoutFormatDict holds all available LayoutFormats.
//
//nolint:gochecknoglobals // read-only enum dictionary
var LayoutFormatDict = struct {
	GoFormat      LayoutFormat
	UnixTimestamp LayoutFormat
}{
	GoFormat:      LayoutFormatGo,
	UnixTimestamp: LayoutFormatUnixTimestamp,
}

const (
	LayoutTimestampSeconds      = "U@"
	LayoutTimestampMilliseconds = "U@000"
	LayoutTimestampMicroseconds = "U@000000"
	LayoutTimestampNanoseconds  = "U@000000000"
)

// LayoutDetails stores parsed meta information about a layout string such as
// "2006-01-02".
type LayoutDetails struct {
	// MinimalUnit e.g. Day for "2006-01-02" and Month for "2006-01"
	MinimalUnit DateUnit

	// Format is the format of the time used in the layout
	Format LayoutFormat

	// Units met in layout
	Units []DateUnit
}

func (lm *LayoutDetails) HasUnit(q DateUnit) bool {
	return slices.Contains(lm.Units, q)
}

// ParseLayout parses given layout string and returns LayoutDetails, or nil
// when it finds no date unit. Units are found by substring heuristics, so an
// unusual layout can be misread.
func ParseLayout(layout string) *LayoutDetails {
	result := &LayoutDetails{Units: make([]DateUnit, 0)}

	// Day of the month: "2" "_2" "02"
	// weak check for now via regex: 2 not followed by 0 because of 2006
	// TODO: stronger check: 2 should not be part of 2006
	//       also `_2` should not be confused with __2
	twoNotFollowedByZero := regexp.MustCompile(`2([^0]|$)`)
	containsDay := strings.Contains(layout, "_2") || strings.Contains(layout, "02") ||
		twoNotFollowedByZero.MatchString(layout)

	if containsDay {
		result.MinimalUnit = Day
		result.Units = append(result.Units, Day)
	}

	// Reference for future:
	//	Day of the week: "Mon" "Monday"
	//	Day of the year: "__2" "002"

	// Month: "Jan" "January" "01" "1"
	oneNotFollowedByFive := regexp.MustCompile(`0?1([^5]|$)`) // `1` is month but `15` are hours
	containsMonth := strings.Contains(layout, "01") ||
		strings.Contains(layout, "Jan") ||
		oneNotFollowedByFive.MatchString(layout)

	if containsMonth {
		if !result.MinimalUnit.Defined() {
			result.MinimalUnit = Month
		}
		result.Units = append(result.Units, Month)
	}

	// Year: "2006" "06"
	containsYear := strings.Contains(layout, "2006") || strings.Contains(layout, "06")
	if containsYear {
		if !result.MinimalUnit.Defined() {
			result.MinimalUnit = Year
		}
		result.Units = append(result.Units, Year)
	}

	if strings.Contains(layout, LayoutTimestampSeconds) {
		result.Format = LayoutFormatUnixTimestamp

		switch {
		case strings.Contains(layout, LayoutTimestampNanoseconds):
			result.Units = append(result.Units, UnixNanosecond)
			result.MinimalUnit = UnixNanosecond
		case strings.Contains(layout, LayoutTimestampMicroseconds):
			result.Units = append(result.Units, UnixMicrosecond)
			result.MinimalUnit = UnixMicrosecond
		case strings.Contains(layout, LayoutTimestampMilliseconds):
			result.Units = append(result.Units, UnixMillisecond)
			result.MinimalUnit = UnixMillisecond
		default:
			result.Units = append(result.Units, UnixSecond)
			result.MinimalUnit = UnixSecond
		}

		return result
	}

	if len(result.Units) == 0 {
		return nil
	}

	result.Format = LayoutFormatGo
	return result
}

// findTimestampPart returns the [start, end) byte range of the timestamp token
// in layout, e.g. `foobar_U@000.log` gives 7, 12. It returns 0, 0 when the
// layout has none.
func findTimestampPart(layout string) (int, int) {
	if !strings.Contains(layout, LayoutTimestampSeconds) {
		return 0, 0
	}

	var start, end int
	switch {
	case strings.Contains(layout, LayoutTimestampNanoseconds):
		start = strings.Index(layout, LayoutTimestampNanoseconds)
		end = start + len(LayoutTimestampNanoseconds)
	case strings.Contains(layout, LayoutTimestampMicroseconds):
		start = strings.Index(layout, LayoutTimestampMicroseconds)
		end = start + len(LayoutTimestampMicroseconds)
	case strings.Contains(layout, LayoutTimestampMilliseconds):
		start = strings.Index(layout, LayoutTimestampMilliseconds)
		end = start + len(LayoutTimestampMilliseconds)
	default:
		start = strings.Index(layout, LayoutTimestampSeconds)
		end = start + len(LayoutTimestampSeconds)
	}

	return start, end
}

// layoutWildcard at the start or end of a Go layout matches any run of
// characters, including none: "Mon Jan 02 2006 15:04:05 GMT-0700*".
const layoutWildcard = "*"

var errMisplacedWildcard = errors.New("* is allowed only at the start or end of a layout")

// wildcardLayout is a Go layout with its edge wildcards split off.
type wildcardLayout struct {
	core              string
	leading, trailing bool
}

// maxMatchLen bounds the text the core layout can match, so the wildcard
// search never tries wider candidates. A Go layout token reads at most twice
// its width ("1" reads "12", "Monday" reads "Wednesday"); the slack covers
// zone names and the fractional seconds Go accepts after a seconds field.
func (w wildcardLayout) maxMatchLen() int {
	const slack = 32
	return 2*len(w.core) + slack
}

// splitWildcards strips one layoutWildcard from each edge of layout. It
// reports false when a wildcard is left in the core, or the core is empty.
func splitWildcards(layout string) (wildcardLayout, bool) {
	var w wildcardLayout
	w.core, w.leading = strings.CutPrefix(layout, layoutWildcard)
	w.core, w.trailing = strings.CutSuffix(w.core, layoutWildcard)
	if w.core == "" || strings.Contains(w.core, layoutWildcard) {
		return wildcardLayout{}, false
	}
	return w, true
}
