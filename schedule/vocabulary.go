package schedule

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

// Weekday sets behind the built-in day-set terms.
//
//nolint:gochecknoglobals // read-only weekday sets
var (
	weekendDays  = []time.Weekday{time.Saturday, time.Sunday}
	workweekDays = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}
)

// Vocabulary resolves human terms ("weekends", "mornings") to week patterns.
// The zero-config default (package-level ParsePatterns) uses
// DefaultDayPartition and the built-in weekday terms; apps may build their
// own to add domain words or a custom partition.
//
// It is separate from the root package's aliases, which resolve to absolute
// instants or periods ("last-week" is a concrete week); a week pattern is a
// recurring predicate with no anchor in absolute time.
type Vocabulary struct {
	Partition DayPartition
	Extra     map[string]WeekPatterns // app-specific terms, e.g. "officehours"
}

// DefaultVocabulary is the built-in vocabulary: weekday terms plus
// DefaultDayPartition part names.
func DefaultVocabulary() Vocabulary {
	return Vocabulary{Partition: DefaultDayPartition}
}

// ParsePatterns parses s using DefaultVocabulary.
func ParsePatterns(s string) (WeekPatterns, error) {
	return DefaultVocabulary().ParsePatterns(s)
}

// ParsePatterns converts a comma-separated list of vocabulary terms into a
// union of week patterns:
//
//	input  := term ("," term)*      // comma = union (OR)
//	term   := dayset | daypart
//	dayset := "monday(s)" … "sunday(s)" | "weekend(s)"
//	        | "workingday(s)" | "workday(s)" | "weekday(s)"
//	daypart:= partition part name, singular or plural
//
// Terms are case-insensitive, whitespace-trimmed and plural-tolerant. The
// result is a list because mixed unions ("sundays,mornings" = sundays all
// day OR mornings every day) are not expressible in a single {Days, Spans}
// conjunction. An unknown term errors listing the known vocabulary.
func (v Vocabulary) ParsePatterns(s string) (WeekPatterns, error) {
	if strings.TrimSpace(s) == "" {
		return nil, errors.New("empty patterns input")
	}

	table := v.terms()

	var patterns WeekPatterns
	for term := range strings.SplitSeq(s, ",") {
		term = strings.ToLower(strings.TrimSpace(term))
		if term == "" {
			return nil, fmt.Errorf("empty term in patterns input %q", s)
		}

		resolved, ok := table[term]
		if !ok {
			// Plural tolerance is a lookup detail: "sundays" resolves via "sunday".
			resolved, ok = table[strings.TrimSuffix(term, "s")]
		}
		if !ok {
			return nil, unknownTermError(term, table)
		}

		patterns = append(patterns, resolved...)
	}

	return patterns, nil
}

// terms builds the full term table: weekday names, built-in day-set words,
// the partition's part names, and any Extra entries. Keys are canonical
// (singular, lowercase); plural tolerance is handled at lookup time, not by
// storing N entries. Keep it a data table: it is where a localized vocabulary
// would plug in.
func (v Vocabulary) terms() map[string]WeekPatterns {
	table := make(map[string]WeekPatterns)

	for _, day := range allWeekdays {
		table[strings.ToLower(day.String())] = WeekPatterns{{Days: []time.Weekday{day}}}
	}

	table["weekend"] = WeekPatterns{{Days: slices.Clone(weekendDays)}}
	for _, synonym := range []string{"workingday", "workday", "weekday"} {
		table[synonym] = WeekPatterns{{Days: slices.Clone(workweekDays)}}
	}

	for _, part := range v.Partition {
		table[strings.ToLower(part.Name)] = WeekPatterns{{Spans: []DaySlot{part.Span}}}
	}

	for extraTerm, extraPatterns := range v.Extra {
		table[strings.ToLower(extraTerm)] = extraPatterns
	}

	return table
}

// unknownTermError lists the known vocabulary: these errors reach end users
// (HTTP 400s), so they must say what would have worked.
func unknownTermError(term string, table map[string]WeekPatterns) error {
	known := slices.Sorted(maps.Keys(table))
	if strings.Contains(term, "-") {
		return fmt.Errorf(
			"unknown pattern term %q: combined day-part terms are not supported yet (known terms: %s)",
			term, strings.Join(known, ", "))
	}
	return fmt.Errorf("unknown pattern term %q (known terms: %s)", term, strings.Join(known, ", "))
}
