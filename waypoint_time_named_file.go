package years

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TimeNamedWaypointFile is a Waypoint implementation for files/directories
type TimeNamedWaypointFile struct {
	*WaypointFile

	// timeInput is the full input that is using for time.Parse()
	// Usually it's considered to be file/dir name + names of parents (if necessary)
	timeInput string

	// layout is a full layout (knowing required parent's layout information)
	// e.g. "2006/Jan"
	layout string

	// Unit of waypoint representing the duration unit (day|month|year)
	unit DateUnit
}

func (w *TimeNamedWaypointFile) setNonCalendar() {
	w.layout = ""
	w.timeInput = ""
	w.t = time.Time{}
}

type TimeNamedWaypointFiles []*TimeNamedWaypointFile

func (w *TimeNamedWaypointFile) Time() time.Time { return w.t }

func NewTimeNamedWaypointFile(path string, fullLayout string, parentArg ...*TimeNamedWaypointFile) (*TimeNamedWaypointFile, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	w0 := &WaypointFile{path: path, fileInfo: stat, t: stat.ModTime()}
	w := &TimeNamedWaypointFile{WaypointFile: w0}

	fullLayoutParts := strings.Split(fullLayout, string(os.PathSeparator))
	layout := fullLayoutParts[0] // by default layout would be first part of layout parts

	w.timeInput = w.fileInfo.Name()

	if len(parentArg) > 0 && parentArg[0] != nil {
		parent := parentArg[0]
		ownLayout := strings.TrimPrefix(fullLayout, parent.layout+"/")

		if w.fileInfo.IsDir() {
			ownLayout = strings.Split(ownLayout, string(os.PathSeparator))[0]
		}

		layout = parent.layout + string(os.PathSeparator) + ownLayout

		if parent.timeInput != "" {
			w.timeInput = parent.timeInput + string(os.PathSeparator) + w.timeInput
		}
	}

	layout = strings.TrimPrefix(layout, string(os.PathSeparator))
	w.layout = layout

	// Default parser is used. Use years.SetParserDefaults to configure parsing
	w.t, err = NewParser().Parse(layout, w.timeInput)
	if err != nil {
		w.setNonCalendar()
	} else {
		layoutDetails := ParseLayout(layout)
		w.unit = layoutDetails.MinimalUnit
	}

	if w.fileInfo.IsDir() {

		// Go deeper in the directory
		innerPaths, err := filepath.Glob(filepath.Join(w.path, "*"))
		if err != nil {
			return nil, err
		}

		for _, innerPath := range innerPaths {
			child, err := NewTimeNamedWaypointFile(innerPath, fullLayout, w)
			if err != nil {
				// TODO(nice-to-have): add configurable way to halt on child error, to log/omit errors, etc
				log.Printf("child: NewTimeNamedWaypointFile(%s) failed: %s\n", innerPath, err)
				continue
			}

			// By default, let's sort nodes in Past->Future order
			var inserted bool
			for i, existed := range w.waypoints {
				if existed.Time().After(child.t) {
					w.waypoints = append(w.waypoints[:i+1], w.waypoints[i:]...)
					w.waypoints[i] = child
					inserted = true
					break
				}
			}
			if !inserted {
				w.waypoints = append(w.waypoints, child)
			}
		}
	}

	return w, nil
}
