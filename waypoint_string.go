package years

import (
	"time"
)

// WaypointString is a Waypoint implementation for a string that represents time.
type WaypointString struct {
	// timeInput is the full input passed to Parse.
	timeInput string

	// layout is the layout to parse timeInput with; empty means the default
	// parser's layouts.
	layout string

	t time.Time
}

type WaypointStrings []*WaypointString

// NewWaypointString parses v with the optional layout, or the parser's layouts
// when none is given. A v that does not parse becomes a non-calendar waypoint
// with a zero time and an empty identifier.
func NewWaypointString(v string, layoutArg ...string) *WaypointString {
	w := &WaypointString{timeInput: v}
	if len(layoutArg) > 0 {
		w.layout = layoutArg[0]
	}

	var err error

	// NewParser() takes the package defaults; configure them with SetParserDefaults.
	w.t, err = NewParser().Parse(w.layout, w.timeInput)
	if err != nil {
		w.setNonCalendar()
	}

	return w
}

func (w *WaypointString) Time() time.Time                       { return w.t }
func (w *WaypointString) Identifier() string                    { return w.timeInput }
func (w *WaypointString) IsContainer() bool                     { return false }
func (w *WaypointString) Children() []Waypoint                  { return nil }
func (w *WaypointString) Voyager(parserArg ...*Parser) *Voyager { return NewVoyager(w, parserArg...) }

func (w *WaypointString) setNonCalendar() {
	w.layout = ""
	w.timeInput = ""
	w.t = time.Time{}
}

func WaypointsFromStrings(timeStrings []string, layoutArg ...string) []Waypoint {
	ws := make([]Waypoint, len(timeStrings))
	for i, v := range timeStrings {
		ws[i] = NewWaypointString(v, layoutArg...)
	}
	return ws
}

func WaypointGroupFromStrings(timeStrings []string, layoutArg ...string) Waypoint {
	return NewWaypointGroup("", WaypointsFromStrings(timeStrings, layoutArg...)...)
}
