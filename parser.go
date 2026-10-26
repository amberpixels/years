package years

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type ParserOption func(*Parser)

type Parser struct {
	acceptUnixSeconds bool
	acceptUnixMilli   bool
	acceptUnixMicro   bool
	acceptUnixNano    bool

	acceptAliases bool

	acceptRelativeOffsets bool

	clock   Clock
	layouts []string

	// location is where offset-less layouts are read and aliases resolve; nil
	// keeps layouts in UTC and aliases in the clock's zone.
	location *time.Location

	weekStart time.Weekday
	workdays  [daysInWeek]bool

	// aliases is the resolved table: core aliases for this parser's config with
	// customAliases on top. NewParser builds it once all options have run.
	aliases       map[string]func(time.Time) time.Time
	customAliases map[string]func(time.Time) time.Time
}

func WithLayouts(layouts ...string) ParserOption {
	return func(p *Parser) { p.layouts = append(p.layouts, layouts...) }
}

func AcceptUnixSeconds() ParserOption { return func(p *Parser) { p.acceptUnixSeconds = true } }
func AcceptUnixMilli() ParserOption   { return func(p *Parser) { p.acceptUnixMilli = true } }
func AcceptUnixMicro() ParserOption   { return func(p *Parser) { p.acceptUnixMicro = true } }
func AcceptUnixNano() ParserOption    { return func(p *Parser) { p.acceptUnixNano = true } }

func AcceptAliases() ParserOption {
	return func(p *Parser) { p.acceptAliases = true }
}

// AcceptRelativeOffsets opts in to duration-offset expressions ("12h", "7d",
// "2w", "3mo", "1y"). An offset denotes the period of the last N units (see
// ParsePeriod); the instant Parse returns that period's start, i.e. now - offset.
func AcceptRelativeOffsets() ParserOption {
	return func(p *Parser) { p.acceptRelativeOffsets = true }
}

// WithCustomAliases merges the given aliases into the parser's alias table
// (custom entries win over same-named core ones). The caller's map is copied,
// never mutated, and the option may be given more than once.
func WithCustomAliases(customAliases map[string]func(time.Time) time.Time) ParserOption {
	return func(p *Parser) {
		merged := maps.Clone(p.customAliases)
		if merged == nil {
			merged = make(map[string]func(time.Time) time.Time, len(customAliases))
		}
		maps.Copy(merged, customAliases)
		p.customAliases = merged
	}
}

// WithWeekStart sets the weekday the week-based aliases ("this-week",
// "last-week", "next-week") start on. Defaults to time.Sunday. A value outside
// Sunday..Saturday leaves the week start unchanged.
func WithWeekStart(d time.Weekday) ParserOption {
	return func(p *Parser) {
		if d >= time.Sunday && d <= time.Saturday {
			p.weekStart = d
		}
	}
}

// WithWorkdays sets which weekdays are working days, used by the
// "last-workday" alias. Defaults to Monday through Friday. The given days
// replace the set; called with no valid days it leaves the set unchanged,
// as an empty working week has no last working day.
func WithWorkdays(days ...time.Weekday) ParserOption {
	return func(p *Parser) {
		var set [daysInWeek]bool
		var found bool
		for _, d := range days {
			if d >= time.Sunday && d <= time.Saturday {
				set[d] = true
				found = true
			}
		}
		if found {
			p.workdays = set
		}
	}
}

// WithLocation reads layouts that carry no offset in loc, as
// time.ParseInLocation does, and resolves aliases and relative offsets on
// loc's calendar. Values with an explicit offset keep it. A nil loc leaves the
// setting unchanged.
func WithLocation(loc *time.Location) ParserOption {
	return func(p *Parser) {
		if loc != nil {
			p.location = loc
		}
	}
}

// WithCustomClock opts to enable a custom Clock.
func WithCustomClock(c Clock) ParserOption {
	return func(p *Parser) { p.clock = c }
}

// defaultParserOptions are what NewParser applies when given no options.
//
//nolint:gochecknoglobals // package-wide parser defaults
var defaultParserOptions []ParserOption

// DefaultLayouts are the unambiguous Go layouts the default parser accepts, so
// years.JustParse reads RFC3339 timestamps and plain dates with no setup. They
// run most specific first because the first match wins. RFC3339Nano stands
// alone: its trailing-.9 form matches with or without a fractional second.
//
//nolint:gochecknoglobals // exported, read-only set of layout constants
var DefaultLayouts = []string{
	time.RFC3339Nano,
	LayoutDateTime,
	LayoutDateTimeShort,
	LayoutDate,
}

func ResetParserDefaults() {
	SetParserDefaults(
		AcceptUnixSeconds(),
		AcceptAliases(),
		AcceptRelativeOffsets(),
		WithLayouts(DefaultLayouts...),
	)
}

//nolint:gochecknoinits // package-level Parse needs the defaults set before first use
func init() { ResetParserDefaults() }

func GetParserDefaults() []ParserOption      { return defaultParserOptions }
func SetParserDefaults(opts ...ParserOption) { defaultParserOptions = opts }
func ExtendParserDefaults(opts ...ParserOption) {
	defaultParserOptions = append(defaultParserOptions, opts...)
}

func NewParser(options ...ParserOption) *Parser {
	p := &Parser{
		clock:     stdClock,
		weekStart: time.Sunday,
		workdays:  defaultWorkdays(),
	}

	if len(options) == 0 {
		options = defaultParserOptions
	}

	for _, opt := range options {
		opt(p)
	}

	// Built after the options so week start and working days apply regardless
	// of option order, and custom aliases still win over core ones.
	if p.weekStart == time.Sunday && p.workdays == defaultWorkdays() && len(p.customAliases) == 0 {
		p.aliases = defaultCoreAliases
	} else {
		p.aliases = coreAliasesFor(p.weekStart, p.workdays)
		maps.Copy(p.aliases, p.customAliases)
	}

	return p
}

// DefaultParser builds a fresh parser from the current package defaults, so
// SetParserDefaults takes effect on the next call.
//
//nolint:gochecknoglobals // replaceable constructor
var DefaultParser = func() *Parser {
	return NewParser(defaultParserOptions...)
}

// ParseEpoch reads v as a unix timestamp in every enabled unit and keeps the
// readings that land in [1970, 3000). It returns the coarsest one (seconds,
// then milli, micro, nano) and reports whether more than one unit fit; enable
// a single unit when values near 1970 must not be ambiguous.
func (p *Parser) ParseEpoch(v int64) (time.Time, bool, error) {
	if !p.acceptUnixSeconds && !p.acceptUnixMilli && !p.acceptUnixMicro && !p.acceptUnixNano {
		return time.Time{}, false, errors.New("no units enabled")
	}

	const (
		secToMilli = 1_000
		secToMicro = 1_000_000
		secToNano  = 1_000_000_000
	)

	// Plausible window: [1970-01-01, 3000-01-01).
	timestampMin := time.Unix(0, 0)                             // 1970-01-01
	timestampMax := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC) // 3000-01-01

	type cand struct {
		t   time.Time
		src string
	}

	var candidates []cand

	if p.acceptUnixSeconds {
		if c := time.Unix(v, 0).UTC(); c.Before(timestampMax) && !c.Before(timestampMin) {
			candidates = append(candidates, cand{c, "sec"})
		}
	}

	if p.acceptUnixMilli {
		ns := v * (secToNano / secToMilli)
		if c := time.Unix(0, ns).UTC(); c.Before(timestampMax) && !c.Before(timestampMin) {
			candidates = append(candidates, cand{c, "milli"})
		}
	}
	if p.acceptUnixMicro {
		ns := v * (secToNano / secToMicro)
		if c := time.Unix(0, ns).UTC(); c.Before(timestampMax) && !c.Before(timestampMin) {
			candidates = append(candidates, cand{c, "micro"})
		}
	}
	if p.acceptUnixNano {
		if c := time.Unix(0, v).UTC(); c.Before(timestampMax) && !c.Before(timestampMin) {
			candidates = append(candidates, cand{c, "nano"})
		}
	}

	switch len(candidates) {
	case 0:
		return time.Time{}, false, errors.New("timestamp out of plausible range for all allowed units")
	case 1:
		return candidates[0].t, false, nil
	default:
		return candidates[0].t, true, nil
	}
}

// Parse parses value with layout, or with each of the parser's layouts when
// layout is empty; a given layout is strict and its mismatch is an error. An
// all-digit value is tried as an epoch first when a unix unit is accepted;
// accepted aliases and relative offsets are tried after the layouts.
func (p *Parser) Parse(layout, value string) (time.Time, error) {
	digits, parseIntErr := strconv.ParseInt(value, 10, 64)
	isNumericValue := parseIntErr == nil

	if isNumericValue {
		if p.acceptUnixSeconds || p.acceptUnixMilli || p.acceptUnixMicro || p.acceptUnixNano {
			parsedEpoch, _, err := p.ParseEpoch(digits)
			if err == nil {
				return parsedEpoch, nil
			}

			// TODO: fail here when no layout can match an all-digit value; the
			// layouts below are left to reject it.
		} else if len(p.layouts) == 0 {
			return time.Time{}, errors.New("misconfiguration")
		}
	}

	layouts := p.layouts
	var strictLayout bool
	if layout != "" {
		strictLayout = true
		layouts = []string{layout}
	}
	for _, l := range layouts {
		layoutDetails := ParseLayout(l)
		if layoutDetails == nil {
			// No date unit in the layout, e.g. a bare "*".
			return time.Time{}, fmt.Errorf("unknown layout format: %s", l)
		}

		switch layoutDetails.Format {
		case LayoutFormatGo:
			if t, err := p.parseGoLayout(l, value); err == nil {
				return t, nil
			} else if strictLayout {
				return time.Time{}, fmt.Errorf("failed to parse time with layout(%s): %w", l, err)
			}
		case LayoutFormatUnixTimestamp:
			start, end := findTimestampPart(l)
			if start == 0 && end == 0 {
				continue
			}
			if len(value) < end || len(value) < start {
				continue
			}

			afterTimestamp := layout[end:]
			beforeTimestamp := layout[:start]

			// Strip the layout's literal prefix and suffix, leaving the digits.
			cleanValue := value
			cleanValue = strings.TrimPrefix(cleanValue, beforeTimestamp)
			cleanValue = strings.TrimSuffix(cleanValue, afterTimestamp)

			cleanDigits, err := strconv.ParseInt(cleanValue, 10, 64)
			if err == nil {
				parsedEpoch, _, err := p.ParseEpoch(cleanDigits)
				return parsedEpoch, err
			} else if strictLayout {
				return time.Time{}, fmt.Errorf("failed to parse time with layout(%s): %w", l, err)
			}
		case LayoutFormatUndefined:
			return time.Time{}, fmt.Errorf("unknown layout format: %s", l)
		default:
			return time.Time{}, fmt.Errorf("unknown layout format: %s", l)
		}
	}

	if p.acceptAliases {
		for alias, aliasCb := range p.aliases {
			if value == alias {
				return aliasCb(p.now()), nil
			}
		}
	}

	if p.acceptRelativeOffsets {
		if period, ok := parseRelativeOffset(value, p.now()); ok {
			return period.Start, nil
		}
	}

	return time.Time{}, errors.New("unable to parse time")
}

// JustParse is a shortcut for Parse("", value) (so using all parser's accepted layouts).
func (p *Parser) JustParse(value string) (time.Time, error) {
	return p.Parse("", value)
}

// now is the clock's time, moved into the parser's location when one is set.
func (p *Parser) now() time.Time {
	if p.location != nil {
		return p.clock.Now().In(p.location)
	}
	return p.clock.Now()
}

// parseGoLayout parses value with a Go layout that may start and/or end with
// a wildcard. A trailing wildcard tries the longest prefix of value first and a
// leading one the shortest skip, so variable-width fields keep every digit;
// with both, the longest substring wins, leftmost first.
func (p *Parser) parseGoLayout(layout, value string) (time.Time, error) {
	w, ok := splitWildcards(layout)
	if !ok {
		return time.Time{}, errMisplacedWildcard
	}
	if !w.leading && !w.trailing {
		return p.parseGoLayoutExact(layout, value)
	}

	// Error of the whole value against the core: it names the offending text.
	t, firstErr := p.parseGoLayoutExact(w.core, value)
	if firstErr == nil {
		return t, nil
	}

	n := len(value)
	isCut := func(i int) bool { return i == n || utf8.RuneStart(value[i]) }
	for length := n - 1; length > 0; length-- {
		for start := 0; start+length <= n; start++ {
			end := start + length
			if (start > 0 && !w.leading) || (end < n && !w.trailing) {
				continue
			}
			if !isCut(start) || !isCut(end) {
				continue
			}
			if t, err := p.parseGoLayoutExact(w.core, value[start:end]); err == nil {
				return t, nil
			}
		}
	}
	return time.Time{}, firstErr
}

// parseGoLayoutExact parses value with a Go layout, in the parser's location
// when one is set and in UTC otherwise. A date-only value in the location
// starts at the day's first instant.
func (p *Parser) parseGoLayoutExact(layout, value string) (time.Time, error) {
	if p.location == nil {
		return time.Parse(layout, value)
	}
	t, err := time.ParseInLocation(layout, value, p.location)
	if err == nil && layoutIsDateOnly(layout) && (t.Hour() != 0 || t.Minute() != 0) {
		// The date's midnight falls in a DST gap and was moved into the day before.
		y, m, d := t.Date()
		t = startOfDay(y, m, d+1, t.Location())
	}
	return t, err
}
