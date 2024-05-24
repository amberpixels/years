package years

import (
	"regexp"
	"strings"
)

// DateUnit stays for the unit of a date like Day/Month/Year
// DateTime units (like Hour, Minute, etc) are not supported (they are not needed)
// TODO: consider supporting week and quarter
type DateUnit int

const (
	UnitUndefined DateUnit = iota
	// Day as day of the month
	// TODO support day of the week + day of the year
	Day DateUnit = 1 << (iota - 1)
	Month
	Year
)

func (du DateUnit) String() string {
	switch du {
	case Day:
		return "day"
	case Month:
		return "month"
	case Year:
		return "year"
	case UnitUndefined:
		return ""
	default:
		panic("fix DateUnit enum!")
	}
}

func (du DateUnit) Defined() bool { return du != UnitUndefined }

// DateUnitsDict holds all available DateUnits
var DateUnitsDict = struct {
	Day   DateUnit
	Month DateUnit
	Year  DateUnit
}{
	Day:   Day,
	Month: Month,
	Year:  Year,
}

// LayoutMeta stores parsed meta information about given layout string
// e.g. "2006-02-01"
type LayoutMeta struct {
	// MinimalUnit e.g. Day for "2006-01-02" and Month for "2006-01"
	MinimalUnit DateUnit

	// GoFormat is true when layout is in pure Go time.Time layout format
	// Currently it's always true
	// TODO: support formats that are popular in JS (YYYY, etc)
	GoFormat bool

	// Units met in layout
	Units []DateUnit
}

func (lm *LayoutMeta) HasUnit(q DateUnit) bool {
	for _, u := range lm.Units {
		if u == q {
			return true
		}
	}
	return false
}

func (lm *LayoutMeta) HasYear() bool  { return lm.HasUnit(Year) }
func (lm *LayoutMeta) HasMonth() bool { return lm.HasUnit(Month) }
func (lm *LayoutMeta) HasDay() bool   { return lm.HasUnit(Day) }

// parseLayout returns one of the units: year/month/day
// by the given format
// Note: it's a pretty hacky/weak function, but we're OK with it for now
func parseLayout(layout string) *LayoutMeta {
	result := &LayoutMeta{Units: make([]DateUnit, 0)}

	// Day of the month: "2" "_2" "02"
	// weak check for now via regex: 2 not followed by 0 because of 2006
	// TODO: stronger check: 2 should not be part of 2006
	//       also `_2` should not be confused with __2
	twoNotFollowedByZero := regexp.MustCompile(`2([^0]|$)`)
	containsDay := strings.Contains(layout, "_2") || strings.Contains(layout, "02") || twoNotFollowedByZero.MatchString(layout)
	if containsDay {
		result.MinimalUnit = Day
		result.Units = append(result.Units, Day)
	}

	// Reference for future:
	//	Day of the week: "Mon" "Monday"
	//	Day of the year: "__2" "002"

	// Month: "Jan" "January" "01" "1"
	oneNotFollowedByFive := regexp.MustCompile(`0?1([^5]|$)`) // `1` is month but `15` are hours
	containsMonth := strings.Contains(layout, "01") || strings.Contains(layout, "Jan") || oneNotFollowedByFive.MatchString(layout)
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

	if len(result.Units) == 0 {
		// temporary very simple way of saying it's not a valid layout
		return nil
	}

	// for now all layouts are Go-format only
	result.GoFormat = true

	return result
}
