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
// week start and working days.
// Aliases that are timezone-dependent by default use timezone of given base time
//
// TODO(nice-to-have): refactor keys are not just hardcoded strings,
// but should be language-depended, so they can be translated.
func coreAliasesFor(weekStart time.Weekday, workdays [daysInWeek]bool) map[string]func(time.Time) time.Time {
	return map[string]func(time.Time) time.Time{
		"today": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToDay().Time()
		},
		"yesterday": func(base time.Time) time.Time {
			base = base.AddDate(0, 0, -1)
			return Mutate(&base).TruncateToDay().Time()
		},
		"tomorrow": func(base time.Time) time.Time {
			base = base.AddDate(0, 0, 1)
			return Mutate(&base).TruncateToDay().Time()
		},
		"this-week": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToWeek(weekStart).Time()
		},
		"last-week": func(base time.Time) time.Time {
			base = base.AddDate(0, 0, -daysInWeek)
			return Mutate(&base).TruncateToWeek(weekStart).Time()
		},
		"next-week": func(base time.Time) time.Time {
			base = base.AddDate(0, 0, daysInWeek)
			return Mutate(&base).TruncateToWeek(weekStart).Time()
		},
		// to avoid misunderstanding we deliberately do not have `this-weekend` alias
		// as it can be considered as both "following weekend" or "previous weekend"
		aliasNextWeekend: func(base time.Time) time.Time {
			followingSaturday := base
			for followingSaturday.Weekday() != time.Saturday {
				followingSaturday = followingSaturday.AddDate(0, 0, 1)
			}
			return Mutate(&followingSaturday).TruncateToDay().Time()
		},
		aliasLastWeekend: func(base time.Time) time.Time {
			lastSaturday := base
			for lastSaturday.Weekday() != time.Saturday {
				lastSaturday = lastSaturday.AddDate(0, 0, -1)
			}
			lastSunday := lastSaturday.AddDate(0, 0, -1)
			return Mutate(&lastSunday).TruncateToDay().Time()
		},
		// the most recent working day strictly before base's day. Each midnight is
		// built from the calendar date: AddDate on a truncated time keeps drifting
		// off 00:00 once the walk crosses a midnight DST change.
		aliasLastWorkday: func(base time.Time) time.Time {
			y, m, d := base.Date()
			var day time.Time
			for i := 1; i <= daysInWeek; i++ {
				day = time.Date(y, m, d-i, 0, 0, 0, 0, base.Location())
				if workdays[day.Weekday()] {
					break
				}
			}
			return day
		},
		"this-month": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToMonth().Time()
		},
		"last-month": func(base time.Time) time.Time {
			// Operate on the 1st so the month step is overflow-safe
			// (AddDate on e.g. Mar 31 would otherwise spill into Feb).
			startOfMonth := Mutate(&base).TruncateToMonth().Time()
			return startOfMonth.AddDate(0, -1, 0)
		},
		"next-month": func(base time.Time) time.Time {
			startOfMonth := Mutate(&base).TruncateToMonth().Time()
			return startOfMonth.AddDate(0, 1, 0)
		},
		"this-year": func(base time.Time) time.Time {
			return Mutate(&base).TruncateToYear().Time()
		},
		"last-year": func(base time.Time) time.Time {
			startOfYear := Mutate(&base).TruncateToYear().Time()
			return startOfYear.AddDate(-1, 0, 0)
		},
		"next-year": func(base time.Time) time.Time {
			startOfYear := Mutate(&base).TruncateToYear().Time()
			return startOfYear.AddDate(1, 0, 0)
		},
	}
}
