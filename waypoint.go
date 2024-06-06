package years

import "time"

// Waypoint is an interface for objects that have a time.
type Waypoint interface {
	// Identifier returns the identifier of the object.
	// E.g. for file waypoints it can be file path.
	Identifier() string

	// Time returns the time of the object.
	Time() time.Time

	// IsContainer returns true if the object can contain other objects.
	// E.g. for directories, it should return true.
	IsContainer() bool

	// Children returns the children of the object if it's a container.
	// E.g. for directories, it should return the list of files and directories inside.
	Children() []Waypoint
}

func NewWaypointGroup(identifier string, waypoints ...Waypoint) Waypoint {
	return &WaypointGroup{identifier: identifier, waypoints: waypoints}
}

type WaypointGroup struct {
	waypoints  []Waypoint
	identifier string
}

func (wg *WaypointGroup) Time() time.Time      { return time.Time{} }
func (wg *WaypointGroup) Identifier() string   { return "" }
func (wg *WaypointGroup) IsContainer() bool    { return true }
func (wg *WaypointGroup) Children() []Waypoint { return wg.waypoints }

func AllChildren(w Waypoint) []Waypoint {
	var result []Waypoint
	if w.IsContainer() {
		for _, child := range w.Children() {
			result = append(result, child)
			result = append(result, AllChildren(child)...)
		}
	}

	return result
}
