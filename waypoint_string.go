package years

import (
	"time"
)

// WaypointString is a Waypoint implementation for a string that represents time
type WaypointString struct {
	// timeInput is the full input that is using for time.Parse()
	timeInput string

	// layout is a layout to be used for parsing time
	// If not specified, default parser will be used (with default layouts)
	layout string

	// t is the time of the waypoint
	t time.Time
}

func (w *WaypointString) setNonCalendar() {
	w.layout = ""
	w.timeInput = ""
	w.t = time.Time{}
}

type WaypointStrings []*WaypointString

func (w *WaypointString) Time() time.Time                       { return w.t }
func (w *WaypointString) Identifier() string                    { return w.timeInput }
func (w *WaypointString) IsContainer() bool                     { return false }
func (w *WaypointString) Children() []Waypoint                  { return nil }
func (w *WaypointString) Voyager(parserArg ...*Parser) *Voyager { return NewVoyager(w, parserArg...) }

func NewWaypointString(v string, layoutArg ...string) *WaypointString {
	w := &WaypointString{timeInput: v}
	if len(layoutArg) > 0 {
		w.layout = layoutArg[0]
	}

	var err error

	// Default parser is used. Use years.SetParserDefaults to configure parsing
	w.t, err = NewParser().Parse(w.layout, w.timeInput)
	if err != nil {
		w.setNonCalendar()
	}

	return w
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
