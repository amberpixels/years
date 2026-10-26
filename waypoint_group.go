package years

import "time"

// WaypointGroup is a Waypoint that only contains other waypoints.
type WaypointGroup struct {
	waypoints  []Waypoint
	identifier string
}

// NewWaypointGroup creates a group for given waypoints.
func NewWaypointGroup(identifier string, waypoints ...Waypoint) Waypoint {
	return &WaypointGroup{identifier: identifier, waypoints: waypoints}
}

// Time returns the zero time: a group has no time of its own.
// TODO(nice-to-have): make it configurable, e.g. the min, max or first child's time.
func (wg *WaypointGroup) Time() time.Time      { return time.Time{} }
func (wg *WaypointGroup) Identifier() string   { return wg.identifier }
func (wg *WaypointGroup) IsContainer() bool    { return true }
func (wg *WaypointGroup) Children() []Waypoint { return wg.waypoints }
