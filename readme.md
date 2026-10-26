<p align="center">
  <img src="logo.svg" alt="years" width="380">
</p>

<div align="center">

### Everything has a time. Travel it.

A time toolkit for Go: parsing, formatting, humanizing, and navigating time-based data.

[![Go Reference](https://pkg.go.dev/badge/github.com/amberpixels/years.svg)](https://pkg.go.dev/github.com/amberpixels/years)
[![Go Version](https://img.shields.io/github/go-mod/go-version/amberpixels/years)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-yellow.svg)](LICENSE)

</div>

---

`years` bundles the time helpers you keep rewriting. One parser accepts Go layouts, unix timestamps, and human aliases. Around it: zero-safe formatting with canonical layouts, "3h ago" humanization, fluent in-place mutation, and a Voyager that traverses anything with a time, such as calendar-structured file trees.

```go
t, _ := years.JustParse("2024-05-26") // ISO dates & RFC3339 out of the box
t, _ = years.JustParse("1717852417")  // unix timestamps
t, _ = years.JustParse("yesterday")   // human aliases

years.Humanize(t) // "1d ago"
```

## Install

```sh
go get github.com/amberpixels/years
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/amberpixels/years"
)

func main() {
	created, _ := years.JustParse("2024-05-26 13:45")

	fmt.Println(years.Format(created, years.LayoutHumanDate)) // "May 26, 2024"
	fmt.Println(years.Humanize(created))                      // "2y ago"

	years.Mutate(&created).TruncateToDay().SetYear(2025)
	fmt.Println(years.Format(created, years.LayoutDate)) // "2025-05-26"

	d, _ := years.ParseISODuration("PT1H2M3S")
	fmt.Println(years.FormatDurationClock(d)) // "1:02:03"
}
```

## Parsing

With no setup, the default parser reads RFC3339 timestamps, `2006-01-02 15:04[:05]`, plain dates, unix-second timestamps, and aliases:

```go
t, _ := years.JustParse("2024-05-26T13:45:00Z")
t, _ = years.JustParse("1717852417")
t, _ = years.JustParse("next-week")

// Strict single-layout parsing, like time.Parse:
t, _ = years.Parse("2006-01-02", "2024-05-26")
```

Aliases resolve against the package clock: `today`, `yesterday`, `tomorrow`, `last-workday`, `this-week`, `last-week`, `next-week`, `last-weekend`, `next-weekend`, `this-month`, `last-month`, `next-month`, `this-year`, `last-year`, `next-year`.

`last-workday` is the most recent working day before today: on a Monday, that is Friday. Weeks start on Sunday and working days are Monday to Friday; `WithWeekStart` and `WithWorkdays` change them.

Layouts may also contain a timestamp part: `U@` for unix seconds, `U@000` for milliseconds, `U@000000` / `U@000000000` for micro/nanoseconds:

```go
p := years.NewParser(
	years.AcceptUnixMilli(),
	years.AcceptAliases(),
	years.WithLayouts("2006", "2006-01", "2006-Jan-02"),
)

t, _ = p.Parse("logs-U@000.log", "logs-1717852417000.log") // 2024-06-08 13:13:37 UTC
t, _ = p.JustParse("2020-01")
```

A Go layout may start and/or end with `*`, which matches any run of characters, including none. It reads dates wrapped in text no layout can describe, such as JavaScript's `Date.toString()`:

```go
t, _ = years.Parse("Mon Jan 02 2006 15:04:05 GMT-0700*",
	"Thu Apr 24 2025 11:52:05 GMT+0300 (Eastern European Summer Time)") // 2025-04-24 11:52:05 +0300
t, _ = years.Parse("*2006-01-02*", "[2025-04-24]")
```

`*` is allowed only as the first or last character. `Parse` with a layout that has `*` anywhere else returns an error; in a `WithLayouts` list such a layout never matches. A trailing `*` keeps the longest prefix that parses and a leading `*` the shortest skip, so `2006-1-2*` reads `2025-4-24 x` as April 24, not April 2. With both, the longest substring that parses wins, leftmost first. Wildcard layouts are lenient, so in a `WithLayouts` list put them after the strict ones.

For the global parser, `SetParserDefaults` replaces its options, `ExtendParserDefaults` appends, and `ResetParserDefaults` restores the built-in behavior:

```go
years.ExtendParserDefaults(years.WithLayouts("2006, Jan 2"))
t, _ = years.JustParse("2020, Dec 1")

// Monday-start weeks and a Sunday-to-Thursday working week:
years.ExtendParserDefaults(
	years.WithWeekStart(time.Monday),
	years.WithWorkdays(time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday),
)
period, _ := years.JustParsePeriod("last-workday")
```

By default, a layout value with no offset is read in UTC (`2026-10-20` is UTC midnight). Aliases and relative offsets resolve in the clock's zone, which is local time for the standard clock. `WithLocation` puts both in one zone: offset-less values are read in it and day and week boundaries are cut in it. Values with an explicit offset (`Z`, `+02:00`) keep their own:

```go
years.ExtendParserDefaults(years.WithLocation(time.Local))
t, _ = years.JustParse("2026-10-20") // local midnight
```

Options passed to `NewParser` replace the defaults instead of adding to them. A custom parser that should resolve aliases lists `AcceptAliases()` alongside `WithWeekStart` or `WithWorkdays`.

## Formatting

Named display layouts, so you stop re-typing reference-time strings, plus zero- and nil-safe helpers:

```go
t := time.Date(2025, time.April, 30, 13, 45, 0, 0, time.UTC)

years.Format(t, years.LayoutDate)          // "2025-04-30"
years.Format(t, years.LayoutDateTime)      // "2025-04-30 13:45:00"
years.Format(t, years.LayoutDateTimeShort) // "2025-04-30 13:45"
years.Format(t, years.LayoutHuman)         // "Apr 30, 2025 13:45"
years.Format(t, years.LayoutHumanDate)     // "Apr 30, 2025"

years.Format(time.Time{}, years.LayoutDate) // "" - the zero time renders empty
years.FormatPtr(nil, years.LayoutDate)      // "" - nil-safe
```

## Durations

Parse ISO-8601 durations (e.g. YouTube's `contentDetails.duration`), and render durations media-clock style or in a compact human form:

```go
d, _ := years.ParseISODuration("PT15M30S") // 15m30s
years.FormatDurationClock(d)               // "15:30"
years.FormatDurationClock(time.Hour + 2*time.Minute + 3*time.Second) // "1:02:03"

years.HumanizeDuration(90 * time.Minute) // "1h 30m"
years.HumanizeDuration(45 * time.Second) // "45s"
```

## Relative Time

```go
years.Humanize(time.Now().Add(-3 * time.Hour)) // "3h ago"
years.Humanize(time.Now().Add(48 * time.Hour)) // "in 2d"

// Clock-free variant (no global clock needed; handy in tests):
years.HumanizeFrom(base, base.Add(-5*time.Minute)) // "5m ago"
```

`Humanize` reads the package clock (`years.Now()`), so tests can make it deterministic via `years.SetStdClock`.

## Mutating Time

`Mutate` wraps a `*time.Time` with fluent setters and truncation helpers that modify it in place:

```go
t := time.Date(2025, time.April, 30, 13, 45, 59, 0, time.UTC)

years.Mutate(&t).TruncateToWeek(time.Monday) // t is now 2025-04-28 00:00:00
years.Mutate(&t).SetMonth(time.December).SetDay(24).SetHour(18)
```

Truncation covers every unit from `TruncateToSecond` to `TruncateToYear`; setters run from `SetYear` to `SetNanosecond`.

## Voyager: Navigating Time-Based Data

Anything implementing `Waypoint` (an identifier, a time, optional children) can be traversed with a `Voyager`. Strings work out of the box:

```go
dates := []string{"2024-03-01", "2024-01-01", "2024-02-01"}

v := years.NewVoyager(years.WaypointGroupFromStrings(dates))
_ = v.Traverse(func(w years.Waypoint) {
	fmt.Println(w.Identifier()) // 2024-03-01, 2024-02-01, 2024-01-01
}, years.O_PAST()) // newest first
```

Pass a non-default layout to the waypoints: `years.WaypointGroupFromStrings(dates, "2006, Jan 2")`.

Files and directories named after dates form navigable calendars. Given a tree like `calendar/2024/Jan/2024-01-15.txt`:

```go
wf, err := years.NewTimeNamedWaypointFile("calendar", "2006/Jan/2006-01-02.txt")
if err != nil {
	panic(err)
}

v := years.NewVoyager(wf)
_ = v.Traverse(func(w years.Waypoint) {
	fmt.Println(w.Identifier()) // file paths, oldest first
}, years.O_FUTURE(), years.O_LEAVES_ONLY())

// Jump straight to a date (aliases work here too):
found, _ := v.Navigate("yesterday")
```

Traverse options: `O_PAST` / `O_FUTURE` for direction, `O_LEAVES_ONLY` / `O_CONTAINERS_ONLY` / `O_ALL` for node filtering, `O_NON_CALENDAR` to include nodes without a parsed time. For time from file metadata (modification/creation/access) instead of names, see `NewWaypointFile`.

## Schedules

The `schedule` subpackage models recurring weekly time windows - working hours, quiet hours, availability:

```go
import "github.com/amberpixels/years/schedule"

work := schedule.Schedule{
	Days:      []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
	StartHour: 9,
	EndHour:   17,
	Gaps:      []schedule.TimeRange{{StartHour: 12, EndHour: 13}}, // lunch break
}

work.Contains(t)         // true on a Monday at 10:30
work.SlotsForDay(t)      // [09:00-12:00, 13:00-17:00]
work.AvailableMinutes(t) // 420
work.NextMatchingDay(t)  // next scheduled day
```

`MultiSlotSchedule` covers disjoint (even cross-midnight) windows per day, and `CompositeSchedule` merges several schedules into one.

### Week patterns

`WeekPattern` is a *recurring* wall-clock predicate such as "weekends" or "nights", with no anchor in absolute time (unlike aliases: `last-week` resolves to one concrete week). `ParsePatterns` turns a human vocabulary into patterns; comma means union:

```go
patterns, _ := schedule.ParsePatterns("weekends,mornings")

patterns.Contains(t) // true on Saturday 15:00, or on any day at 09:30
```

The built-in vocabulary (case-insensitive, plural-tolerant): weekday names (`monday(s)`...), `weekend(s)`, `workingday(s)`/`workday(s)`/`weekday(s)`, and the day parts of `DefaultDayPartition`:

| part | span |
|---|---|
| `night` | 22:00-05:00 (wraps midnight) |
| `morning` | 05:00-12:00 |
| `lunchtime` | 12:00-14:00 |
| `afternoon` | 14:00-17:00 |
| `evening` | 17:00-22:00 |

The partition covers the full day exactly once (enforced by `Validate`), so `PartOf` doubles as a group-by key:

```go
schedule.DefaultDayPartition.PartOf(t).Name // "morning" | "lunchtime" | ...
```

A cross-midnight span belongs to the day it starts on, so Sunday night includes Monday 01:00. Patterns read the given time's wall clock as-is; there is no timezone field, so convert times to the zone you mean first. Add domain words and custom partitions with `Vocabulary`:

```go
vocab := schedule.DefaultVocabulary()
vocab.Extra = map[string]schedule.WeekPatterns{"officehours": {{
	Days:  []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
	Spans: []schedule.DaySlot{{Start: schedule.TimeOfDay{Hour: 9}, End: schedule.TimeOfDay{Hour: 17}}},
}}}

patterns, _ := vocab.ParsePatterns("officehours")
```

## Feedback

`years` is a solo, opinionated project - but if you stumbled upon it and have
ideas, questions, or bug reports, an [issue](https://github.com/amberpixels/years/issues) is always welcome :)

## License

[MIT](LICENSE) © [amberpixels](https://amberpixels.io)
