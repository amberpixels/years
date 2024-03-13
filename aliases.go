package years

import "time"

// TODO: refactor `New()` asap. Pointers are mess.
//       let it be pure functions... with no pointers? or with pointers?
//       as in structs *time.Time is usually used because of nil-check and omitempty (check this)

var builtinTimeAliases = map[string]func(time.Time) time.Time{
	"today": func(base time.Time) time.Time {
		New(&base).TruncateToDay()
		return base
	},
	"yesterday": func(base time.Time) time.Time {
		base = base.AddDate(0, 0, -1)
		New(&base).TruncateToDay()
		return base
	},
	"tomorrow": func(base time.Time) time.Time {
		base = base.AddDate(0, 0, 1)
		New(&base).TruncateToDay()
		return base
	},
	// TODO:
	"this-week": nil,
	"last-week": nil,
	"next-week": nil,

	"this-weekend": nil,
	"last-weekend": nil,

	"this-month": nil,
	"last-month": nil,
	"next-month": nil,

	// todo: quarter

	"this-year": nil,
	"last-year": nil,
	"next-year": nil,
}
