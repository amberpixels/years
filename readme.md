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

`years` bundles the time helpers you keep rewriting: one parser that accepts Go layouts, unix timestamps, and human aliases; zero-safe formatting with canonical layouts; "3h ago" humanization; fluent in-place mutation; and a Voyager for traversing anything that has a time - like calendar-structured file trees.

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

The default parser understands RFC3339 timestamps, `2006-01-02 15:04[:05]`, plain dates, unix-second timestamps, and aliases - no setup:

```go
t, _ := years.JustParse("2024-05-26T13:45:00Z")
t, _ = years.JustParse("1717852417")
t, _ = years.JustParse("next-week")

// Strict single-layout parsing, like time.Parse:
t, _ = years.Parse("2006-01-02", "2024-05-26")
```

Aliases resolve against the package clock: `today`, `yesterday`, `tomorrow`, `this-week`, `last-week`, `next-week`, `last-weekend`, `next-weekend`, `this-month`, `last-month`, `next-month`, `this-year`, `last-year`, `next-year`.

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

The global parser is configurable too - `SetParserDefaults` replaces its options, `ExtendParserDefaults` appends, `ResetParserDefaults` restores the out-of-the-box behavior:

```go
years.ExtendParserDefaults(years.WithLayouts("2006, Jan 2"))
t, _ = years.JustParse("2020, Dec 1")
```

## Formatting

Canonical display layouts (so you stop re-typing the reference-time magic strings), plus zero/nil-safe helpers:

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

Truncation goes down to any unit (`TruncateToSecond` ... `TruncateToYear`), setters cover `SetYear` ... `SetNanosecond`.

## Voyager: Navigating Time-Based Data

Anything implementing `Waypoint` (an identifier, a time, optional children) can be traversed with a `Voyager`. Strings work out of the box:

```go
dates := []string{"2024-03-01", "2024-01-01", "2024-02-01"}

v := years.NewVoyager(years.WaypointGroupFromStrings(dates))
_ = v.Traverse(func(w years.Waypoint) {
	fmt.Println(w.Identifier()) // 2024-03-01, 2024-02-01, 2024-01-01
}, years.O_PAST()) // newest first
```

Non-default layouts are given to the waypoints: `years.WaypointGroupFromStrings(dates, "2006, Jan 2")`.

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

## Contributing

Issues and pull requests are welcome - see [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT - see [LICENSE](LICENSE).
