package years

import "time"

const (
	daysInWeek = 7
)

// Alias names shared between the instant table and the period logic.
const (
	aliasNextWeekend = "next-weekend"
	aliasLastWeekend = "last-weekend"
	aliasLastWorkday = "last-workday"
)

// defaultWorkdays is the working-day set a parser starts with: Monday to Friday.
func defaultWorkdays() [daysInWeek]bool {
	var set [daysInWeek]bool
	for d := time.Monday; d <= time.Friday; d++ {
		set[d] = true
	}
	return set
}

// defaultCoreAliases is the built-in table for the default config, shared by
// every parser that keeps it, so package-level calls do not rebuild it. It is
// never written: a parser with custom aliases builds its own table.
//
//nolint:gochecknoglobals // read-only, built once
var defaultCoreAliases = coreAliasesFor(time.Sunday, defaultWorkdays())

// coreAliasesFor returns a fresh table of the built-in aliases for the given
// week start and working days. Every alias resolves on base's calendar, in
// base's location.
//
// TODO(nice-to-have): make the keys translatable instead of hardcoded English.
func coreAliasesFor(weekStart time.Weekday, workdays [daysInWeek]bool) map[string]func(time.Time) time.Time {
	return map[string]func(time.Time) time.Time{
		"today": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToDay().Time()
		},
		"yesterday": func(base time.Time) time.Time {
			return dayStartFrom(base, -1)
		},
		"tomorrow": func(base time.Time) time.Time {
			return dayStartFrom(base, 1)
		},
		"this-week": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToWeek(weekStart).Time()
		},
		"last-week": func(base time.Time) time.Time {
			day := dayStartFrom(base, -daysInWeek)
			return Mutate(&day).TruncateToWeek(weekStart).Time()
		},
		"next-week": func(base time.Time) time.Time {
			day := dayStartFrom(base, daysInWeek)
			return Mutate(&day).TruncateToWeek(weekStart).Time()
		},
		// There is no this-weekend: it reads as either the coming or the past one.
		aliasNextWeekend: func(base time.Time) time.Time {
			return dayStartFrom(base, daysUntilSaturday(base))
		},
		aliasLastWeekend: func(base time.Time) time.Time {
			return dayStartFrom(base, -daysSinceSaturday(base)-1)
		},
		// The most recent working day strictly before base's day.
		aliasLastWorkday: func(base time.Time) time.Time {
			var day time.Time
			for i := 1; i <= daysInWeek; i++ {
				day = dayStartFrom(base, -i)
				if workdays[day.Weekday()] {
					break
				}
			}
			return day
		},
		"this-month": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToMonth().Time()
		},
		// Built from the 1st so the month step is overflow-safe
		// (Mar 31 minus a month would otherwise spill into Feb).
		"last-month": func(base time.Time) time.Time {
			return startOfDay(base.Year(), base.Month()-1, 1, base.Location())
		},
		"next-month": func(base time.Time) time.Time {
			return startOfDay(base.Year(), base.Month()+1, 1, base.Location())
		},
		"this-year": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToYear().Time()
		},
		"last-year": func(base time.Time) time.Time {
			return startOfDay(base.Year()-1, 1, 1, base.Location())
		},
		"next-year": func(base time.Time) time.Time {
			return startOfDay(base.Year()+1, 1, 1, base.Location())
		},
	}
}

// dayStartFrom returns the start of the day n calendar days from base's day.
// It steps by date, not by AddDate on a wall clock, so a day whose midnight
// falls in a DST gap is never skipped or shifted.
func dayStartFrom(base time.Time, n int) time.Time {
	y, m, d := base.Date()
	return startOfDay(y, m, d+n, base.Location())
}

// daysUntilSaturday counts the days from t's weekday forward to Saturday.
func daysUntilSaturday(t time.Time) int {
	return (int(time.Saturday) - int(t.Weekday()) + daysInWeek) % daysInWeek
}

// daysSinceSaturday counts the days from Saturday forward to t's weekday.
func daysSinceSaturday(t time.Time) int {
	return (int(t.Weekday()) - int(time.Saturday) + daysInWeek) % daysInWeek
}
