package years

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/djherbis/times"
)

// WaypointFile is a Waypoint implementation for files/directories.
type WaypointFile struct {
	// Path of the waypoint
	// It's the request field for .prepare() to be called
	path string

	// fileInfo holds the file info for the given file
	fileInfo os.FileInfo

	// timeSpec holds cross-platform file time creation/modification/access/birth information
	timeSpec times.Timespec

	// t is the time of the waypoint
	t time.Time

	// Waypoints are inner children (subdirectories, files, etc)
	waypoints []Waypoint
}

type WaypointFiles []*WaypointFile

func (w *WaypointFile) Time() time.Time      { return w.t }
func (w *WaypointFile) Identifier() string   { return w.path }
func (w *WaypointFile) IsContainer() bool    { return w.fileInfo.IsDir() }
func (w *WaypointFile) Children() []Waypoint { return w.waypoints }

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
		// Go deeper in the directory
		innerPaths, err := filepath.Glob(filepath.Join(w.path, "*"))
		if err != nil {
			return nil, err
		}

		for _, innerPath := range innerPaths {
			child, err := NewWaypointFile(innerPath, timeGetter)
			if err != nil {
				// TODO(nice-to-have): add configurable way to halt on child error, to log/omit errors, etc
				log.Printf("child: NewWaypointFile(%s) failed: %s\n", innerPath, err)
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
