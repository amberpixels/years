package years

import (
	"errors"
	"fmt"
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
	return func(p *Parser) { p.layouts = layouts }
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
		for k, v := range customAliases {
			p.aliases[k] = v
		}
	}
}

func WithCustomClock(c Clock) ParserOption {
	return func(p *Parser) { p.clock = c }
}

// defaultOptions are applied to the
var defaultOptions = []ParserOption{
	AcceptUnixSeconds(),
	AcceptAliases(),
}

func SetDefaults(opts ...ParserOption) { defaultOptions = opts }
func GetDefaults() []ParserOption      { return defaultOptions }

func NewParser(options ...ParserOption) *Parser {
	p := &Parser{
		clock:   stdClock,
		aliases: coreAliases,
	}

	if len(options) == 0 {
		options = defaultOptions
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

var DefaultParser = func() *Parser {
	return NewParser(defaultOptions...)
}

func (p *Parser) ParseTimestamp(value string) (time.Time, error) {
	unixDigits, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	switch true {
	case p.acceptUnixSeconds:
		return time.Unix(unixDigits, 0), nil
	case p.acceptUnixMilli:
		unixDigits *= int64(time.Millisecond)
	case p.acceptUnixMicro:
		unixDigits *= int64(time.Microsecond)
	case p.acceptUnixNano:
		unixDigits *= int64(time.Nanosecond)
	}
	// TODO(nice-to-have): add more validation here (e.g. check if len of digits is reasonable for milli/micro/nano)

	return time.Unix(0, unixDigits), nil
}

// ParseTimeWithLayout parses time from given value using given layout (or using all accepted layouts if layout is empty)
func (p *Parser) ParseTimeWithLayout(layout string, value string) (time.Time, error) {

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

			if t, err := p.ParseTimestamp(cleanValue); err == nil {
				return t, nil
			} else if strictLayout {
				return time.Time{}, fmt.Errorf("failed to parse time with layout(%s): %w", l, err)
			}
		default:
			return time.Time{}, fmt.Errorf("unknown layout format: %s", l)
		}
	}

	// then try to parse as unix timestamp
	if p.acceptUnixSeconds || p.acceptUnixMilli || p.acceptUnixMicro || p.acceptUnixNano {
		if t, err := p.ParseTimestamp(value); err == nil {
			return t, nil
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

// ParseTime is a shortcut for ParseTimeLayout("", value) (so all layouts are accepted)
func (p *Parser) ParseTime(value string) (time.Time, error) {
	return p.ParseTimeWithLayout("", value)
}
