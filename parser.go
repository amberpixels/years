package years

import (
	"errors"
	"strconv"
	"time"
)

type ParserOption func(*Parser)

type Parser struct {
	acceptUnix      bool
	acceptUnixMilli bool
	acceptAliases   bool

	clock   Clock
	layouts []string

	aliases map[string]func(time.Time) time.Time
}

func WithLayouts(layouts ...string) ParserOption {
	return func(p *Parser) { p.layouts = layouts }
}

func AcceptUnix() ParserOption {
	return func(p *Parser) { p.acceptUnix = true }
}

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

func AcceptUnixMilli() ParserOption {
	return func(p *Parser) { p.acceptUnixMilli = true }
}

func WithCustomClock(c Clock) ParserOption {
	return func(p *Parser) { p.clock = c }
}

// defaultOptions are applied to the
var defaultOptions = []ParserOption{
	AcceptUnix(),
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

func (p *Parser) ParseTime(value string) (time.Time, error) {
	if p.acceptUnixMilli && p.acceptUnix {
		// this is fragile!
		// Better do not use both unix & unix milli
		if len(value) >= 13 {
			if unixMilli, err := strconv.ParseInt(value[:13], 10, 64); err == nil {
				return time.Unix(0, unixMilli*int64(time.Millisecond)), nil
			}
		}
	} else if p.acceptUnixMilli {
		if unixMilli, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Unix(0, unixMilli*int64(time.Millisecond)), nil
		}
	} else if p.acceptUnix {
		if unixSec, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Unix(unixSec, 0), nil
		}
	}

	if p.acceptAliases {
		for alias, aliasCb := range p.aliases {
			if value == alias {
				return aliasCb(p.clock.Now()), nil
			}
		}
	}

	for _, layout := range p.layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("unable to parse time")
}
