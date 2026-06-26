package years

import "time"

const (
	daysInWeek = 7
)

// coreAliases holds all registered aliases
// Aliases that are timezone-dependent by default use timezone of given base time
//
// TODO(nice-to-have): allow to change Sunday/Monday week start via configuration
// TODO(nice-to-have): refactor keys are not just hardcoded strings,
// but should be language-depended, so they can be translated.
//
//nolint:gochecknoglobals // it's ok
var coreAliases = map[string]func(time.Time) time.Time{
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
		return Mutate(&base).TruncateToWeek(time.Sunday).Time()
	},
	"last-week": func(base time.Time) time.Time {
		base = base.AddDate(0, 0, -daysInWeek)
		return Mutate(&base).TruncateToWeek(time.Sunday).Time()
	},
	"next-week": func(base time.Time) time.Time {
		base = base.AddDate(0, 0, daysInWeek)
		return Mutate(&base).TruncateToWeek(time.Sunday).Time()
	},
	// to avoid misunderstanding we deliberately do not have `this-weekend` alias
	// as it can be considered as both "following weekend" or "previous weekend"
	"next-weekend": func(base time.Time) time.Time {
		followingSaturday := base
		for followingSaturday.Weekday() != time.Saturday {
			followingSaturday = followingSaturday.AddDate(0, 0, 1)
		}
		return Mutate(&followingSaturday).TruncateToDay().Time()
	},
	"last-weekend": func(base time.Time) time.Time {
		lastSaturday := base
		for lastSaturday.Weekday() != time.Saturday {
			lastSaturday = lastSaturday.AddDate(0, 0, -1)
		}
		lastSunday := lastSaturday.AddDate(0, 0, -1)
		return Mutate(&lastSunday).TruncateToDay().Time()
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
