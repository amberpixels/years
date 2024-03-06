package years

import (
	"errors"
	"strconv"
	"time"
)

type ParserOption func(*Parser)

type Parser struct {
	layouts         []string
	acceptUnix      bool
	acceptUnixMilli bool
}

func WithLayouts(layouts ...string) ParserOption {
	return func(p *Parser) {
		p.layouts = layouts
	}
}

func AcceptUnix() ParserOption {
	return func(p *Parser) {
		p.acceptUnix = true
	}
}

func AcceptUnixMilli() ParserOption {
	return func(p *Parser) {
		p.acceptUnixMilli = true
	}
}

// defaultOptions are applied to the
var defaultOptions = []ParserOption{
	AcceptUnix(),
}

func SetDefaults(opts ...ParserOption) { defaultOptions = opts }
func GetDefaults() []ParserOption      { return defaultOptions }

func NewParser(options ...ParserOption) *Parser {
	p := &Parser{}

	// Empty parser can parse only UNIX timestamps
	if len(options) == 0 {
		options = []ParserOption{
			AcceptUnix(),
		}
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

	for _, layout := range p.layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("unable to parse time")
}
