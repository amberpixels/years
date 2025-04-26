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

// DigitsAsTimestamp converts given int64 as a time.Time,
// considering input as seconds/milliseconds/microseconds/nanoseconds.
func (p *Parser) DigitsAsTimestamp(unixDigits int64) time.Time {
	switch {
	case p.acceptUnixNano:
		unixDigits *= int64(time.Nanosecond)
	case p.acceptUnixMicro:
		unixDigits *= int64(time.Microsecond)
	case p.acceptUnixMilli:
		unixDigits *= int64(time.Millisecond)
	case p.acceptUnixSeconds:
		return time.Unix(unixDigits, 0)
	}

	// TODO(nice-to-have): add more validation here:
	// e.g. check if len of digits is reasonable for milli/micro/nano.

	return time.Unix(0, unixDigits)
}

// Parse parses time from given value using given layout (or using all parser's accepted layouts if layout is empty).
func (p *Parser) Parse(layout string, value string) (time.Time, error) {
	// Shorthand: if possible, try to parse as a numeric timestamp:
	digits, parseIntErr := strconv.ParseInt(value, 10, 64)
	isNumericValue := parseIntErr == nil

	if isNumericValue {
		if p.acceptUnixSeconds || p.acceptUnixMilli || p.acceptUnixMicro || p.acceptUnixNano {
			return p.DigitsAsTimestamp(digits), nil
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
				return p.DigitsAsTimestamp(cleanDigits), nil
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
