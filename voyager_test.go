package years_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/amberpixels/years"
	"github.com/djherbis/times"
	"github.com/expectto/be"
)

// voyagerSetup is the per-test setup: it pins the package clock to a known
// "now" (a Tuesday in March 2024) and resets the global parser defaults,
// optionally extending them with the layouts a given test needs. The defaults
// are restored when the test finishes so tests stay independent.
func voyagerSetup(t *testing.T, layouts ...string) {
	t.Helper()
	years.SetStdClock(&StaticClock{
		now: time.Date(2024, time.March, 05, 14, 30, 59, 0, time.UTC),
	})
	years.ResetParserDefaults()
	if len(layouts) > 0 {
		years.ExtendParserDefaults(years.WithLayouts(layouts...))
	}
	t.Cleanup(years.ResetParserDefaults)
}

// collectTraverse runs v.Traverse with the given options, collecting every
// visited waypoint's identifier in visit order.
func collectTraverse(t *testing.T, v *years.Voyager, opts ...years.TraverseOption) []string {
	t.Helper()
	identifiers := make([]string, 0)
	err := v.Traverse(func(w years.Waypoint) {
		identifiers = append(identifiers, w.Identifier())
	}, opts...)
	be.Require(t, err).To(be.Succeed())
	return identifiers
}

func TestVoyager_StringWaypoints_SpecificLayout(t *testing.T) {
	voyagerSetup(t)

	newVoyager := func() *years.Voyager {
		return years.NewVoyager(
			years.WaypointGroupFromStrings([]string{
				"2024-03-05",
				"2024-03-06",
				"2024-04-01",
				"2024-03-07",
			}, "2006-01-02"),
		)
	}

	t.Run("Future", func(t *testing.T) {
		identifiers := collectTraverse(t, newVoyager(), years.O_FUTURE())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"2024-03-05",
			"2024-03-06",
			"2024-03-07",
			"2024-04-01",
		}))
	})

	t.Run("Past", func(t *testing.T) {
		identifiers := collectTraverse(t, newVoyager(), years.O_PAST())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"2024-04-01",
			"2024-03-07",
			"2024-03-06",
			"2024-03-05",
		}))
	})
}

func TestVoyager_StringWaypoints_DifferentLayouts(t *testing.T) {
	voyagerSetup(t, "2006-01", "2006-01-02")

	newVoyager := func() *years.Voyager {
		return years.NewVoyager(
			years.WaypointGroupFromStrings([]string{
				"2024-03-05",
				"2024-03-06",
				"2024-03-07",
				"2024-04",
			}),
		)
	}

	t.Run("Future", func(t *testing.T) {
		identifiers := collectTraverse(t, newVoyager(), years.O_FUTURE())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"2024-03-05",
			"2024-03-06",
			"2024-03-07",
			"2024-04",
		}))
	})

	t.Run("Past", func(t *testing.T) {
		identifiers := collectTraverse(t, newVoyager(), years.O_PAST())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"2024-04",
			"2024-03-07",
			"2024-03-06",
			"2024-03-05",
		}))
	})
}

func TestVoyager_FileWaypoints_ByBirthTime(t *testing.T) {
	voyagerSetup(t)
	calendarPath := filepath.Join(TestDataPath, "by_filetime")

	wf, err := years.NewWaypointFile(calendarPath, func(ts times.Timespec) time.Time {
		return ts.BirthTime()
	})
	be.Require(t, err).To(be.Succeed())
	v := years.NewVoyager(wf)

	t.Run("Future / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/by_filetime/first.txt",
			"internal/testdata/by_filetime/foobar/second.txt",
			"internal/testdata/by_filetime/foobar/third.txt",
		}))
	})

	t.Run("Past / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_CONTAINERS_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/by_filetime/foobar",
			"internal/testdata/by_filetime",
		}))
	})

	t.Run("Past / All", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_ALL())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/by_filetime/foobar/third.txt",
			"internal/testdata/by_filetime/foobar/second.txt",
			"internal/testdata/by_filetime/first.txt",
			"internal/testdata/by_filetime/foobar",
			"internal/testdata/by_filetime",
		}))
	})
}

func TestVoyager_FileWaypoints_ByModTime(t *testing.T) {
	voyagerSetup(t)
	calendarPath := filepath.Join(TestDataPath, "by_filetime")

	wf, err := years.NewWaypointFile(calendarPath, func(ts times.Timespec) time.Time {
		return ts.ModTime()
	})
	be.Require(t, err).To(be.Succeed())
	v := years.NewVoyager(wf)

	t.Run("Future / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/by_filetime/first.txt",
			"internal/testdata/by_filetime/foobar/second.txt",
			"internal/testdata/by_filetime/foobar/third.txt",
		}))
	})

	t.Run("Past / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_CONTAINERS_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/by_filetime",
			"internal/testdata/by_filetime/foobar",
		}))
	})

	t.Run("Past / All", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_ALL())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/by_filetime",
			"internal/testdata/by_filetime/foobar",
			"internal/testdata/by_filetime/foobar/third.txt",
			"internal/testdata/by_filetime/foobar/second.txt",
			"internal/testdata/by_filetime/first.txt",
		}))
	})
}

func TestVoyager_TimeNamedFile_Calendar1(t *testing.T) {
	const testCalendarLayout = "2006/Jan/2006-01-02.txt"
	voyagerSetup(t, "2006", "Jan", "2006-01-02")
	calendarPath := filepath.Join(TestDataPath, "calendar1")

	wf, err := years.NewTimeNamedWaypointFile(calendarPath, testCalendarLayout)
	be.Require(t, err).To(be.Succeed())
	v := years.NewVoyager(wf)

	t.Run("traverse Future / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
			"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
			"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
		}))
	})

	t.Run("traverse Future / All nodes", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_ALL())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar1/2024",
			"internal/testdata/calendar1/2024/Jan",
			"internal/testdata/calendar1/2024/Feb",
			"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
			"internal/testdata/calendar1/2024/Mar",
			"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
			"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
		}))
	})

	t.Run("traverse Future / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_CONTAINERS_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar1/2024",
			"internal/testdata/calendar1/2024/Jan",
			"internal/testdata/calendar1/2024/Feb",
			"internal/testdata/calendar1/2024/Mar",
		}))
	})

	t.Run("traverse Past / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
			"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
			"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
		}))
	})

	t.Run("traverse Past / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_CONTAINERS_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar1/2024/Mar",
			"internal/testdata/calendar1/2024/Feb",
			"internal/testdata/calendar1/2024/Jan",
			"internal/testdata/calendar1/2024",
		}))
	})

	t.Run("traverse Past / All nodes", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_ALL())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar1/2024/Mar/2024-03-06.txt",
			"internal/testdata/calendar1/2024/Mar/2024-03-05.txt",
			"internal/testdata/calendar1/2024/Mar",
			"internal/testdata/calendar1/2024/Feb/2024-02-01.txt",
			"internal/testdata/calendar1/2024/Feb",
			"internal/testdata/calendar1/2024/Jan",
			"internal/testdata/calendar1/2024",
		}))
	})

	t.Run("navigate to a specific date", func(t *testing.T) {
		navigated, err := v.Navigate("2024-03-06")
		be.Require(t, err).To(be.Succeed())
		be.Require(t, navigated).NotTo(be.Nil())
		be.Expect(t, navigated.Identifier()).To(
			be.Eq(filepath.Join(calendarPath, "2024", "Mar", "2024-03-06.txt")),
		)
	})

	t.Run("navigate to today", func(t *testing.T) {
		navigated, err := v.Navigate("today")
		be.Require(t, err).To(be.Succeed())
		be.Require(t, navigated).NotTo(be.Nil())
		be.Expect(t, navigated.Identifier()).To(
			be.Eq(filepath.Join(calendarPath, "2024", "Mar", "2024-03-05.txt")),
		)
	})
}

func TestVoyager_TimeNamedFile_Calendar2(t *testing.T) {
	const testCalendarLayout = "2006/Jan/02 Mon.txt"
	// Calendar2 final files are not sufficient for knowing the date on their own
	// (they require parent information), e.g. 2006/Jan/01 Mon.txt.
	// "2006-01-02" is included so the Navigate query ("2024-03-06") can be parsed —
	// the file names themselves use "02 Mon".
	voyagerSetup(t, "2006", "Jan", "02 Mon", "2006-01-02")
	calendarPath := filepath.Join(TestDataPath, "calendar2")

	wf, err := years.NewTimeNamedWaypointFile(calendarPath, testCalendarLayout)
	be.Require(t, err).To(be.Succeed())
	v := years.NewVoyager(wf)

	t.Run("traverse Future / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
			"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
			"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
		}))
	})

	t.Run("traverse Future / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_CONTAINERS_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar2/2024",
			"internal/testdata/calendar2/2024/Feb",
			"internal/testdata/calendar2/2024/Mar",
		}))
	})

	t.Run("traverse Future / All nodes", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_ALL())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar2/2024",
			"internal/testdata/calendar2/2024/Feb",
			"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
			"internal/testdata/calendar2/2024/Mar",
			"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
			"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
		}))
	})

	t.Run("traverse Past / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
			"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
			"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
		}))
	})

	t.Run("traverse Past / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_CONTAINERS_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar2/2024/Mar",
			"internal/testdata/calendar2/2024/Feb",
			"internal/testdata/calendar2/2024",
		}))
	})

	t.Run("traverse Past / All nodes", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_ALL())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/calendar2/2024/Mar/06 Wed.txt",
			"internal/testdata/calendar2/2024/Mar/05 Tue.txt",
			"internal/testdata/calendar2/2024/Mar",
			"internal/testdata/calendar2/2024/Feb/01 Thu.txt",
			"internal/testdata/calendar2/2024/Feb",
			"internal/testdata/calendar2/2024",
		}))
	})

	t.Run("navigate to a specific date", func(t *testing.T) {
		navigated, err := v.Navigate("2024-03-06")
		be.Require(t, err).To(be.Succeed())
		be.Require(t, navigated).NotTo(be.Nil())
		be.Expect(t, navigated.Identifier()).To(
			be.Eq(filepath.Join(calendarPath, "2024", "Mar", "06 Wed.txt")),
		)
	})

	t.Run("navigate to today", func(t *testing.T) {
		navigated, err := v.Navigate("today")
		be.Require(t, err).To(be.Succeed())
		be.Require(t, navigated).NotTo(be.Nil())
		be.Expect(t, navigated.Identifier()).To(
			be.Eq(filepath.Join(calendarPath, "2024", "Mar", "05 Tue.txt")),
		)
	})
}

func TestVoyager_TimeNamedFile_LogsViaTimestamp(t *testing.T) {
	const testCalendarLayout = "foobar_U@000.log"
	voyagerSetup(t)
	calendarPath := filepath.Join(TestDataPath, "logs_via_timestamp")

	wf, err := years.NewTimeNamedWaypointFile(calendarPath, testCalendarLayout)
	be.Require(t, err).To(be.Succeed())
	v := years.NewVoyager(wf)

	t.Run("traverse Future / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/logs_via_timestamp/foobar_1716559191.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559238.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559253.log",
			"internal/testdata/logs_via_timestamp/inner/foobar_1717669999.log",
		}))
	})

	t.Run("traverse Past / Leaves only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_PAST(), years.O_LEAVES_ONLY())
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/logs_via_timestamp/inner/foobar_1717669999.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559253.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559238.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559191.log",
		}))
	})

	t.Run("traverse Future / Containers only", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_CONTAINERS_ONLY())
		// no containers actually
		be.Expect(t, identifiers).To(be.Eq([]string{}))
	})

	t.Run("traverse Future / All", func(t *testing.T) {
		identifiers := collectTraverse(t, v, years.O_FUTURE(), years.O_ALL())
		// no containers - so ALL means same as leaves only
		be.Expect(t, identifiers).To(be.Eq([]string{
			"internal/testdata/logs_via_timestamp/foobar_1716559191.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559238.log",
			"internal/testdata/logs_via_timestamp/foobar_1716559253.log",
			"internal/testdata/logs_via_timestamp/inner/foobar_1717669999.log",
		}))
	})
}
