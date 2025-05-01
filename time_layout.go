package years

import (
	"regexp"
	"slices"
	"strings"
)

// DateUnit stays for the unit of a date like Day/Month/Year/etc.
type DateUnit int

const (
	UnitUndefined DateUnit = iota
	// Day as day of the month
	// TODO(nice-to-have) support day of the week + day of the year.
	Day  DateUnit = 1 << (iota - 1)
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
//nolint:gochecknoglobals // it's ok
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

const (
	LayoutFormatUndefined LayoutFormat = iota
	// LayoutFormatGo is a format that is supported by Go time.Parse.
	LayoutFormatGo LayoutFormat = 1 << (iota - 1)
	// LayoutFormatUnixTimestamp is a format that parses time from Unix timestamp (seconds or milliseconds).
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
		fallthrough
	default:
		panic("fix LayoutFormat enum!")
	}
}

// LayoutFormatDict holds all available LayoutFormats.
//
//nolint:gochecknoglobals // it's ok
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

// LayoutDetails stores parsed meta information about given layout string.
// e.g. "2006-02-01".
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

// ParseLayout parses given layout string and returns LayoutDetails.
//
// Note: it's a pretty hacky/weak function, but we're OK with it for now.
func ParseLayout(layout string) *LayoutDetails {
	result := &LayoutDetails{Units: make([]DateUnit, 0)}

	// Day of the month: "2" "_2" "02"
	// weak check for now via regex: 2 not followed by 0 because of 2006
	// TODO: stronger check: 2 should not be part of 2006
	//       also `_2` should not be confused with __2
	twoNotFollowedByZero := regexp.MustCompile(`2([^0]|$)`)
	containsDay :=
		strings.Contains(layout, "_2") || strings.Contains(layout, "02") ||
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
	containsMonth :=
		strings.Contains(layout, "01") ||
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
		// temporary very simple way of saying it's not a valid layout
		return nil
	}

	// for now, we only support Go-format here
	result.Format = LayoutFormatGo
	return result
}

// find position (start,end) of the timestamp part in the layout (e.g. `U0.` or `U0.000` etc ).
// e.g. `FileName_U0.txt` -> [10, 13].
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
	}

	return start, end
}
