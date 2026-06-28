package years

import "time"

// Canonical display layouts shared across amberpixels apps, centralized here so
// callers stop re-typing the Go reference-time magic strings. These are the
// formatting counterparts to the parser: build a time with Parse/JustParse,
// render it with Format and one of these layouts.
const (
	// LayoutDate is a calendar date: "2006-01-02".
	LayoutDate = "2006-01-02"
	// LayoutDateTime is a date with full wall-clock time: "2006-01-02 15:04:05".
	LayoutDateTime = "2006-01-02 15:04:05"
	// LayoutDateTimeShort is a date with minute precision: "2006-01-02 15:04".
	LayoutDateTimeShort = "2006-01-02 15:04"
	// LayoutHuman is a friendly date+time: "Jan 2, 2006 15:04".
	LayoutHuman = "Jan 2, 2006 15:04"
	// LayoutHumanDate is a friendly date: "Jan 2, 2006".
	LayoutHumanDate = "Jan 2, 2006"
)

// Format renders t using layout, returning "" for the zero time so callers don't
// emit a meaningless "0001-01-01". It is the formatting counterpart to JustParse.
func Format(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

// FormatPtr is the nil-safe variant of Format: a nil (or zero) time yields "".
func FormatPtr(t *time.Time, layout string) string {
	if t == nil {
		return ""
	}
	return Format(*t, layout)
}
