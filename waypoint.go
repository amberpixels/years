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

// AllChildren is a helper function that gets ALL children of a waypoint (recursively)
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
