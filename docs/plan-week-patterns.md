# Plan: week patterns & day partition (`schedule` package)

Status: **shipped** in **v0.3.0**.

## Motivation

Consumers (runwell today, any r3-backed app tomorrow) want to filter and group
records by *recurring wall-clock patterns*: "only sundays", "weekends",
"workingdays", "mornings", "nights". The `schedule` package already models the
underlying semantics — `Days []time.Weekday`, time-of-day windows (`TimeOfDay`,
`DaySlot` with cross-midnight `Hour >= 24`), and `Contains(time.Time)`
predicates — but there is **no string vocabulary** and no normalized value type
a query layer (r3) can introspect and compile to database filters.

This is deliberately different from the existing alias system (`aliases.go`):
aliases resolve to **absolute** instants/periods ("last-week" → a concrete
Monday); a week pattern is a **recurring** predicate with no anchor in absolute
time. The two must not be conflated.

## Goals

- A small, declarative, introspectable value type for recurring weekly
  wall-clock patterns.
- A canonical **day partition**: named, non-overlapping time-of-day parts
  covering the full 24h — one source of truth for both filtering
  (`?when=mornings`) and classification/grouping (`PartOf(t)` for "runs by part
  of day" charts).
- A parser turning a human vocabulary (`"weekends"`, `"mornings"`,
  `"saturdays,sundays"`) into patterns, with an extension seam for
  app-specific vocabularies.
- Zero timezone opinion: patterns match the **wall clock of the given time
  value as-is**. Callers own locality (e.g. by storing local wall-clock
  timestamps).

## Non-goals (v0.3.0)

- Combined day×part grammar (`sunday-mornings`) — the grammar and semantics are
  specified below so nothing painted us into a corner, but parsing support is a
  follow-up.
- Localized vocabularies (existing TODO in `aliases.go` applies here too;
  keep the term tables data-driven so translation slots in later).
- Any r3 awareness. `years` stays dependency-free; r3 imports us, never the
  reverse.

## API design

All new code lives in the existing `schedule` subpackage.

### WeekPattern — the value type

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
for some span. Since spans are capped below 48h, checking `t`'s own date and
the previous date is sufficient. Consequences:

- Plain windows (`nights` over all days) behave as pure minute-of-day
  wraparound: 00:30 and 23:10 both match.
- The future combined `sunday-nights` **includes Monday 01:00** — the night is
  attributed to the day it starts on, matching how humans use the phrase.
  Fixing this rule now means the r3 compiler and any grouping code can rely on
  it and never diverge later.

`Contains` reads `t`'s wall clock via `t.Weekday()`/`t.Hour()`/`t.Minute()` in
`t`'s own location — no `*time.Location` field, unlike `Schedule`. That is the
whole timezone story by design.

Interop nicety (cheap, include): `func (p WeekPattern) MultiSlotSchedule(loc *time.Location) MultiSlotSchedule`
so existing slot math (`SlotsForDay`, `AvailableMinutes`) works on patterns.

### DayPartition — named parts of the day

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
| night     | 22:00–05:00 (wraps midnight) |
| morning   | 05:00–12:00                  |
| lunchtime | 12:00–14:00                  |
| afternoon | 14:00–17:00                  |
| evening   | 17:00–22:00                  |

`PartOf` is total (the invariant guarantees every minute belongs to exactly one
part) — that's what makes it a safe group-by key for charts. A custom partition
that violates the invariant fails `Validate` loudly instead of silently
misclassifying.

### Vocabulary — string → patterns

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

- Case-insensitive, whitespace-trimmed; both singular and plural accepted,
  canonical form is stored (plural-tolerance is a lookup detail, not N map
  entries).
- Unknown term → error that **lists the known vocabulary** (these strings
  surface in HTTP 400s; make them helpful).
- Terms are table-driven (one map built from weekday names + partition +
  `Extra`), no switch statements — this is the translation seam the existing
  `aliases.go` TODO asks for.
- Reserved for the future: `-` as the dayset×daypart combinator
  (`sunday-mornings` → `{Days: [Sun], Spans: [morning]}`). No current term
  contains a hyphen, so the grammar extends without breakage. Until
  implemented, a hyphenated compound errors as unknown (with a "not yet
  supported" hint).

Union across mixed term kinds is why `ParsePatterns` returns `WeekPatterns`
(list = OR) rather than merging into one struct: `"sundays,mornings"` is
*sundays-all-day OR mornings-every-day*, which a single `{Days, Spans}`
conjunction cannot express. Group = AND, list = OR — same shape r3 composes
with `And`/`Or`, so compilation is mechanical.

## Bug fix rider

`WithCustomAliases` (`parser.go:50`) copies in the wrong direction —
`maps.Copy(customAliases, p.aliases)` mutates the caller's map and never
installs anything on the parser. Fix to merge into `p.aliases` (clone first;
the shared `coreAliases` map must never be mutated) + regression test. Ship in
the same release since consumers may reach for custom aliases alongside the
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
