package years

import (
	"time"
)

// WaypointTimeString is a Waypoint implementation for a string that represents time
type WaypointTimeString struct {
	// timeInput is the full input that is using for time.Parse()
	timeInput string

	// layout is a layout to be used for parsing time
	// If not specified, default parser will be used (with default layouts)
	layout string

	// t is the time of the waypoint
	t time.Time
}

func (w *WaypointTimeString) setNonCalendar() {
	w.layout = ""
	w.timeInput = ""
	w.t = time.Time{}
}

type WaypointTimeStrings []*WaypointTimeString

func (w *WaypointTimeString) Time() time.Time      { return w.t }
func (w *WaypointTimeString) Identifier() string   { return w.timeInput }
func (w *WaypointTimeString) IsContainer() bool    { return false }
func (w *WaypointTimeString) Children() []Waypoint { return nil }

func NewWaypointTimeString(v string, layoutArg ...string) *WaypointTimeString {
	w := &WaypointTimeString{timeInput: v}
	if len(layoutArg) > 0 {
		w.layout = layoutArg[0]
	}
	// TODO: un-hardcode layouts here.
	p := NewParser(WithLayouts("2006-01-02", "2006-01"))

	var err error
	w.t, err = p.ParseTimeWithLayout(w.layout, w.timeInput)
	if err != nil {
		w.setNonCalendar()
	}

	return w
}
