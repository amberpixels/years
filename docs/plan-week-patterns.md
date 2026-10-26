# Plan: week patterns & day partition (`schedule` package)

Status: **shipped** in **v0.3.0**.

## Motivation

Consumers (runwell today, any r3-backed app tomorrow) want to filter and group
records by *recurring wall-clock patterns*: "only sundays", "weekends",
"workingdays", "mornings", "nights". The `schedule` package already models the
semantics: `Days []time.Weekday`, time-of-day windows (`TimeOfDay`, `DaySlot`
with cross-midnight `Hour >= 24`), and `Contains(time.Time)` predicates. It has
**no string vocabulary**, and no normalized value type a query layer (r3) can
introspect and compile to database filters.

This is separate from the alias system (`aliases.go`). Aliases resolve to
**absolute** instants or periods ("last-week" is a concrete week); a week
pattern is a **recurring** predicate with no anchor in absolute time. Do not
conflate the two.

## Goals

- A small, introspectable value type for recurring weekly wall-clock patterns.
- A canonical **day partition**: named, non-overlapping time-of-day parts
  covering the full 24h. One table serves both filtering (`?when=mornings`) and
  grouping (`PartOf(t)` for "runs by part of day" charts).
- A parser turning a human vocabulary (`"weekends"`, `"mornings"`,
  `"saturdays,sundays"`) into patterns, with an extension seam for
  app-specific vocabularies.
- Zero timezone opinion: patterns match the **wall clock of the given time
  value as-is**. Callers own locality (e.g. by storing local wall-clock
  timestamps).

## Non-goals (v0.3.0)

- Combined day-and-part grammar (`sunday-mornings`). Its grammar and semantics
  are specified below so the design leaves room for it; parsing is a follow-up.
- Localized vocabularies (the TODO in `aliases.go` applies here too). Keep the
  term tables data-driven so translations can be added.
- Any r3 awareness. `years` stays dependency-free; r3 imports us, never the
  reverse.

## API design

All new code lives in the existing `schedule` subpackage.

### WeekPattern: the value type

```go
// WeekPattern is a recurring weekly wall-clock pattern: a conjunction of
// "weekday ∈ Days" and "time-of-day ∈ one of Spans".
//   - Days empty  => every day
//   - Spans empty => the whole day
// Spans reuse DaySlot; End.Hour may exceed 24 (cross-midnight, attributed to
// the starting day: {22:00, 29:00} means "22:00 through 05:00 next day").
type WeekPattern struct {
	Days  []time.Weekday
	Spans []DaySlot
}

// WeekPatterns is a union (OR) of patterns.
type WeekPatterns []WeekPattern

func (p WeekPattern) Contains(t time.Time) bool
func (ps WeekPatterns) Contains(t time.Time) bool
func (p WeekPattern) Validate() error   // sane hours/minutes, End > Start, End < 48h
```

**`Contains` semantics (attribution rule).** `p.Contains(t)` is true iff there
exists a day `d` (the calendar date of `t` or the previous date) such that
`weekday(d) ∈ Days` (or `Days` is empty) and `t ∈ [d+span.Start, d+span.End)`
for some span. Spans are capped below 48h, so checking `t`'s own date and the
previous date is enough. Consequences:

- Plain windows (`nights` over all days) wrap by minute of day: 00:30 and
  23:10 both match.
- The future combined `sunday-nights` **includes Monday 01:00**: a night
  belongs to the day it starts on, as people use the phrase. The r3 compiler
  and any grouping code rely on this rule, so it must not change.

`Contains` reads `t`'s wall clock via `t.Weekday()`/`t.Hour()`/`t.Minute()` in
`t`'s own location. Unlike `Schedule`, there is no `*time.Location` field, and
no other timezone handling.

Also include `func (p WeekPattern) MultiSlotSchedule(loc *time.Location) MultiSlotSchedule`
so existing slot math (`SlotsForDay`, `AvailableMinutes`) works on patterns.

### DayPartition: named parts of the day

```go
// DayPart is one named slice of the day within a DayPartition.
type DayPart struct {
	Name string  // canonical singular: "morning", "night", ...
	Span DaySlot // may cross midnight (End.Hour >= 24)
}

// DayPartition is an ordered set of DayParts that must cover the full 24h
// exactly once (no gaps, no overlaps, modulo midnight wrap).
type DayPartition []DayPart

func (dp DayPartition) Validate() error       // enforces the invariant above
func (dp DayPartition) PartOf(t time.Time) DayPart // classification for group-by
```

Default partition (exported var, replaceable; `Validate`d in tests):

| part      | span                         |
|-----------|------------------------------|
| night     | 22:00-05:00 (wraps midnight) |
| morning   | 05:00-12:00                  |
| lunchtime | 12:00-14:00                  |
| afternoon | 14:00-17:00                  |
| evening   | 17:00-22:00                  |

`PartOf` is total, since the invariant puts every minute in exactly one part,
which makes it a safe group-by key for charts. A custom partition that breaks
the invariant fails `Validate` instead of misclassifying.

### Vocabulary: string to patterns

```go
// Vocabulary resolves human terms to week patterns. The zero-config default
// (package-level ParsePatterns) uses DefaultDayPartition and the built-in
// weekday terms; apps may build their own to add domain words or a custom
// partition.
type Vocabulary struct {
	Partition DayPartition
	Extra     map[string]WeekPatterns // app-specific terms, e.g. "officehours"
}

func DefaultVocabulary() Vocabulary
func (v Vocabulary) ParsePatterns(s string) (WeekPatterns, error)

// Package-level convenience over DefaultVocabulary().
func ParsePatterns(s string) (WeekPatterns, error)
```

Grammar (v0.3.0):

```
input  := term ("," term)*      // comma = union (OR)
term   := dayset | daypart
dayset := "monday(s)" … "sunday(s)"
        | "weekend(s)"                          -> Sat, Sun
        | "workingday(s)" | "workday(s)" | "weekday(s)" -> Mon–Fri
daypart:= partition name, singular or plural    -> {Days: nil, Spans: [part.Span]}
```

- Case-insensitive and whitespace-trimmed. Singular and plural are both
  accepted; only the canonical form is stored, and plurals are handled at
  lookup, not as extra map entries.
- An unknown term errors with **the known vocabulary listed**. These strings
  surface in HTTP 400s, so make them helpful.
- Terms are table-driven, one map built from weekday names, the partition and
  `Extra`, with no switch statements. This is where translation plugs in, as
  the `aliases.go` TODO asks.
- Reserved for the future: `-` as the dayset-and-daypart combinator
  (`sunday-mornings` gives `{Days: [Sun], Spans: [morning]}`). No current term
  contains a hyphen, so adding it breaks nothing. Until then, a hyphenated
  compound errors as unknown, with a "not yet supported" hint.

`ParsePatterns` returns `WeekPatterns` (list = OR) instead of merging into one
struct, because mixed terms need a union: `"sundays,mornings"` is
*sundays-all-day OR mornings-every-day*, which one `{Days, Spans}` conjunction
cannot express. Group = AND and list = OR is the shape r3 composes with
`And`/`Or`, so compiling to it is mechanical.

## Bug fix rider

`WithCustomAliases` (`parser.go:50`) copies in the wrong direction:
`maps.Copy(customAliases, p.aliases)` mutates the caller's map and installs
nothing on the parser. Fix it to merge into `p.aliases`, cloning first since
the shared `coreAliases` map must never be mutated, and add a regression test.
Ship it in the same release, as consumers may use custom aliases alongside the
new vocabulary.

## File plan

| file | change |
|------|--------|
| `schedule/pattern.go` (new) | `WeekPattern`, `WeekPatterns`, `Contains`, `Validate`, `MultiSlotSchedule` interop |
| `schedule/partition.go` (new) | `DayPart`, `DayPartition`, `Validate`, `PartOf`, `DefaultDayPartition` |
| `schedule/vocabulary.go` (new) | `Vocabulary`, term tables, `ParsePatterns` |
| `schedule/pattern_test.go`, `partition_test.go`, `vocabulary_test.go` (new) | see below |
| `parser.go` | `WithCustomAliases` fix |
| `readme.md` | extend the Schedules section: patterns, partition, vocabulary (lowercase readme is the authoritative one) |

## Testing

- `Contains` table tests: midnight boundaries (21:59/22:00/04:59/05:00 for
  nights), week wrap (Sat 23:30 vs Sun 00:30), attribution rule for a
  hand-built day+wrapping-span pattern, empty Days / empty Spans.
- `DayPartition.Validate`: default passes; gap, overlap, and wrapped-overlap
  partitions fail with precise messages. `PartOf` totality: every minute of a
  synthetic day maps to exactly one part (loop 0..1439).
- Vocabulary: every documented term round-trips; plural/singular; comma
  unions; unknown-term error content; custom `Extra` term; custom partition.
- `WithCustomAliases` regression: custom alias resolves, core aliases intact,
  caller's map not mutated.

## Release

Tag **v0.3.0** (new API, pre-1.0 minor bump). No breaking changes. r3 and
runwell bump to it in their own stages.
