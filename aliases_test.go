package years_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/expectto/be"
)

func TestCoreAliases(t *testing.T) {
	// Use a known Wednesday for predictable week calculations
	base := time.Date(2025, time.May, 7, 15, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name     string
		alias    string
		expected time.Time
	}{
		{"today", "today",
			time.Date(2025, time.May, 7, 0, 0, 0, 0, time.UTC)},
		{"yesterday", "yesterday",
			time.Date(2025, time.May, 6, 0, 0, 0, 0, time.UTC)},
		{"tomorrow", "tomorrow",
			time.Date(2025, time.May, 8, 0, 0, 0, 0, time.UTC)},
		{"this-week", "this-week",
			// Sunday of current week
			time.Date(2025, time.May, 4, 0, 0, 0, 0, time.UTC)},
		{"last-week", "last-week",
			time.Date(2025, time.April, 27, 0, 0, 0, 0, time.UTC)},
		{"next-week", "next-week",
			time.Date(2025, time.May, 11, 0, 0, 0, 0, time.UTC)},
		{"next-weekend", "next-weekend",
			time.Date(2025, time.May, 10, 0, 0, 0, 0, time.UTC)},
		{"last-weekend", "last-weekend",
			time.Date(2025, time.May, 2, 0, 0, 0, 0, time.UTC)},
		{"this-month", "this-month",
			time.Date(2025, time.May, 1, 0, 0, 0, 0, time.UTC)},
		{"last-month", "last-month",
			time.Date(2025, time.April, 1, 0, 0, 0, 0, time.UTC)},
		{"next-month", "next-month",
			time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC)},
		{"this-year", "this-year",
			time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"last-year", "last-year",
			time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)},
		{"next-year", "next-year",
			time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fn, exists := years.CoreAliases[tc.alias]
			be.Require(t, exists).To(be.True(), fmt.Sprintf("alias %q should be registered", tc.alias))
			be.Expect(t, fn(base)).To(be.Eq(tc.expected), fmt.Sprintf("alias %q", tc.alias))
		})
	}
}
