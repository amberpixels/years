package years

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"
)

type ParserOption func(*Parser)

type Parser struct {
	acceptUnixSeconds bool
	acceptUnixMilli   bool
	acceptUnixMicro   bool
	acceptUnixNano    bool

	acceptAliases bool

	clock   Clock
	layouts []string

	aliases map[string]func(time.Time) time.Time
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

func WithCustomAliases(customAliases map[string]func(time.Time) time.Time) ParserOption {
	return func(p *Parser) {
		maps.Copy(customAliases, p.aliases)
	}
}

// WithCustomClock opts to enable a custom Clock.
func WithCustomClock(c Clock) ParserOption {
	return func(p *Parser) { p.clock = c }
}

// defaultParserOptions are applied always by default.
//
//nolint:gochecknoglobals // it's ok
var defaultParserOptions []ParserOption

func ResetParserDefaults() {
	SetParserDefaults(
		AcceptUnixSeconds(),
		AcceptAliases(),
	)
}

//nolint:gochecknoinits // we're fine for now
func init() { ResetParserDefaults() }

func GetParserDefaults() []ParserOption      { return defaultParserOptions }
func SetParserDefaults(opts ...ParserOption) { defaultParserOptions = opts }
func ExtendParserDefaults(opts ...ParserOption) {
	defaultParserOptions = append(defaultParserOptions, opts...)
}

func NewParser(options ...ParserOption) *Parser {
	p := &Parser{
		clock:   stdClock,
		aliases: coreAliases,
	}

	if len(options) == 0 {
		options = defaultParserOptions
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

// DefaultParser makes a default parser
//
//nolint:gochecknoglobals // it's ok
var DefaultParser = func() *Parser {
	return NewParser(defaultParserOptions...)
}

// ParseEpoch converts the given epoch timestamp int64 as a time.Time,
// considering input as seconds/milliseconds/microseconds/nanoseconds.
// Better use one specific configuration: seconds or milliseconds, etc.
// In case if multiple configurations are enabled, there are edge-cases (both seconds/milli from 1970).
func (p *Parser) ParseEpoch(v int64) (time.Time, bool, error) {
	// sanity: at least one unit must be enabled
	if !p.acceptUnixSeconds && !p.acceptUnixMilli && !p.acceptUnixMicro && !p.acceptUnixNano {
		return time.Time{}, false, errors.New("no units enabled")
	}

	const (
		secToMilli = 1_000
		secToMicro = 1_000_000
		secToNano  = 1_000_000_000
	)

	// plausible window [1970-01-01 … 3000-01-01).
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
		ns := int64(v) * (secToNano / secToMilli)
		if c := time.Unix(0, ns); c.Before(timestampMax) && !c.Before(timestampMin) {
			candidates = append(candidates, cand{c, "milli"})
		}
	}
	if p.acceptUnixMicro {
		ns := v * (secToNano / secToMicro)
		if c := time.Unix(0, ns); c.Before(timestampMax) && !c.Before(timestampMin) {
			candidates = append(candidates, cand{c, "micro"})
		}
	}
	if p.acceptUnixNano {
		if c := time.Unix(0, v); c.Before(timestampMax) && !c.Before(timestampMin) {
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

// Parse parses time from given value using given layout (or using all parser's accepted layouts if layout is empty).
func (p *Parser) Parse(layout string, value string) (time.Time, error) {
	// Shorthand: if possible, try to parse as a numeric timestamp:
	digits, parseIntErr := strconv.ParseInt(value, 10, 64)
	isNumericValue := parseIntErr == nil

	if isNumericValue {
		if p.acceptUnixSeconds || p.acceptUnixMilli || p.acceptUnixMicro || p.acceptUnixNano {
			parsedEpoch, _, err := p.ParseEpoch(digits)
			return parsedEpoch, err
		}

		if len(p.layouts) == 0 {
			return time.Time{}, errors.New("misconfiguration")
		}
	}

	// Try to parse time using all accepted layouts
	layouts := p.layouts
	var strictLayout bool
	if layout != "" {
		strictLayout = true
		layouts = []string{layout}
	}
	for _, l := range layouts {
		layoutDetails := ParseLayout(l)

		switch layoutDetails.Format {
		case LayoutFormatGo:
			if t, err := time.Parse(l, value); err == nil {
				return t, nil
			} else if strictLayout {
				return time.Time{}, fmt.Errorf("failed to parse time with layout(%s): %w", l, err)
			}
		case LayoutFormatUnixTimestamp:
			// Extract timestamp part from the layout if any
			start, end := findTimestampPart(l)
			if start == 0 && end == 0 {
				continue
			}
			if len(value) < end || len(value) < start {
				continue
			}

			afterTimestamp := layout[end:]
			beforeTimestamp := layout[:start]

			// store cleaned value (with only timestamp part)
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
			fallthrough
		default:
			return time.Time{}, fmt.Errorf("unknown layout format: %s", l)
		}
	}

	if p.acceptAliases {
		for alias, aliasCb := range p.aliases {
			if value == alias {
				return aliasCb(p.clock.Now()), nil
			}
		}
	}

	return time.Time{}, errors.New("unable to parse time")
}

// JustParse is a shortcut for Parse("", value) (so using all parser's accepted layouts).
func (p *Parser) JustParse(value string) (time.Time, error) {
	return p.Parse("", value)
}
