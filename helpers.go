package years

import "fmt"

func ToWaypoints(ws ...any) []Waypoint {
	waypoints := make([]Waypoint, len(ws), len(ws))
	for i, w := range ws {
		if wp, ok := w.(Waypoint); !ok {
			panic(fmt.Sprintf("not a waypoint: %v", w))
		} else {
			waypoints[i] = wp
		}
	}
	return waypoints
}
