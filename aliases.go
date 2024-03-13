package years

import "time"

// coreAliases holds all registered aliases
// TODO: allow to change Sunday/Monday week start via configuration
var coreAliases = map[string]func(time.Time) time.Time{
	"today": func(base time.Time) time.Time {
		return Wrap(&base).TruncateToDay().Time()
	},
	"yesterday": func(base time.Time) time.Time {
		base = base.AddDate(0, 0, -1)
		return Wrap(&base).TruncateToDay().Time()
	},
	"tomorrow": func(base time.Time) time.Time {
		base = base.AddDate(0, 0, 1)
		return Wrap(&base).TruncateToDay().Time()
	},
	"this-week": func(base time.Time) time.Time {
		startOfWeek := base.AddDate(0, 0, -int(base.Weekday()))
		return Wrap(&startOfWeek).TruncateToDay().Time()
	},
	"last-week": func(base time.Time) time.Time {
		startOfLastWeek := base.AddDate(0, 0, -7-int(base.Weekday()))
		return Wrap(&startOfLastWeek).TruncateToDay().Time()
	},
	"next-week": func(base time.Time) time.Time {
		startOfNextWeek := base.AddDate(0, 0, 7-int(base.Weekday()))
		return Wrap(&startOfNextWeek).TruncateToDay().Time()
	},
	// to avoid misunderstanding we deliberately do not have `this-weekend` alias
	// as it can be considered as both "following weekend" or "previous weekend"
	"next-weekend": func(base time.Time) time.Time {
		followingSaturday := base
		for followingSaturday.Weekday() != time.Saturday {
			followingSaturday = followingSaturday.AddDate(0, 0, 1)
		}
		return Wrap(&followingSaturday).TruncateToDay().Time()
	},
	"last-weekend": func(base time.Time) time.Time {
		lastSaturday := base
		for lastSaturday.Weekday() != time.Saturday {
			lastSaturday = lastSaturday.AddDate(0, 0, -1)
		}
		lastSunday := lastSaturday.AddDate(0, 0, -1)
		return Wrap(&lastSunday).TruncateToDay().Time()
	},
	"this-month": func(base time.Time) time.Time {
		startOfMonth := time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, base.Location())
		return Wrap(&startOfMonth).TruncateToDay().Time()
	},
	"last-month": func(base time.Time) time.Time {
		startOfLastMonth := time.Date(base.Year(), base.Month()-1, 1, 0, 0, 0, 0, base.Location())
		return Wrap(&startOfLastMonth).TruncateToDay().Time()
	},
	"next-month": func(base time.Time) time.Time {
		startOfNextMonth := time.Date(base.Year(), base.Month()+1, 1, 0, 0, 0, 0, base.Location())
		return Wrap(&startOfNextMonth).TruncateToDay().Time()
	},
	"this-year": func(base time.Time) time.Time {
		startOfYear := time.Date(base.Year(), 1, 1, 0, 0, 0, 0, base.Location())
		return Wrap(&startOfYear).TruncateToDay().Time()
	},
	"last-year": func(base time.Time) time.Time {
		startOfLastYear := time.Date(base.Year()-1, 1, 1, 0, 0, 0, 0, base.Location())
		return Wrap(&startOfLastYear).TruncateToDay().Time()
	},
	"next-year": func(base time.Time) time.Time {
		startOfNextYear := time.Date(base.Year()+1, 1, 1, 0, 0, 0, 0, base.Location())
		return Wrap(&startOfNextYear).TruncateToDay().Time()
	},
}
