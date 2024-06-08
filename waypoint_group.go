package years

import "time"

// WaypointGroup stands for a simple implementation of Waypoint that is a container for other waypoints
type WaypointGroup struct {
	waypoints  []Waypoint
	identifier string
}

// NewWaypointGroup create a group for given waypoints
func NewWaypointGroup(identifier string, waypoints ...Waypoint) Waypoint {
	return &WaypointGroup{identifier: identifier, waypoints: waypoints}
}

// Time returns group's time. For now group itself doesn't have a specific time
// TODO(nice-to-have): this maybe configurable, e.g. no-time/min-time(children)/max-time(children)/time(children[0]), etc
func (wg *WaypointGroup) Time() time.Time      { return time.Time{} }
func (wg *WaypointGroup) Identifier() string   { return wg.identifier }
func (wg *WaypointGroup) IsContainer() bool    { return true }
func (wg *WaypointGroup) Children() []Waypoint { return wg.waypoints }
