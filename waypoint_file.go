package years

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/djherbis/times"
)

// WaypointFile is a Waypoint implementation for files/directories.
type WaypointFile struct {
	// path is the file's path and also its Identifier.
	path string

	fileInfo os.FileInfo

	// timeSpec holds cross-platform file time creation/modification/access/birth information
	timeSpec times.Timespec

	t time.Time

	// waypoints are a directory's entries, sorted oldest first.
	waypoints []Waypoint
}

type WaypointFiles []*WaypointFile

// NewWaypointFile builds a waypoint for path and, for a directory, every
// entry below it. timeGetter picks which file time (birth, modification, ...)
// each waypoint carries. An entry that fails to stat is logged and skipped.
func NewWaypointFile(path string, timeGetter func(timeSpec times.Timespec) time.Time) (*WaypointFile, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not os.Stat file: %w", err)
	}

	timeSpec, err := times.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not times.Stat file: %w", err)
	}

	w := &WaypointFile{path: path, fileInfo: stat, timeSpec: timeSpec, t: timeGetter(timeSpec)}

	if stat.IsDir() {
		innerPaths, err := filepath.Glob(filepath.Join(w.path, "*"))
		if err != nil {
			return nil, err
		}

		for _, innerPath := range innerPaths {
			child, err := NewWaypointFile(innerPath, timeGetter)
			if err != nil {
				// TODO(nice-to-have): add configurable way to halt on child error, to log/omit errors, etc
				slog.Info(fmt.Sprintf("child: NewWaypointFile(%s) failed: %s\n", innerPath, err))
				continue
			}

			// Insert in place so children stay sorted oldest first.
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

func (w *WaypointFile) Time() time.Time      { return w.t }
func (w *WaypointFile) Identifier() string   { return w.path }
func (w *WaypointFile) IsContainer() bool    { return w.fileInfo.IsDir() }
func (w *WaypointFile) Children() []Waypoint { return w.waypoints }
