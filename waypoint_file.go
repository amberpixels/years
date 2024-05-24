package years

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"
)

// WaypointFile is a Waypoint implementation for files/directories
type WaypointFile struct {
	// Path of the waypoint
	// It's the request field for .prepare() to be called
	path string

	// Time representing the start of its range (e.g. start of the day for daily waypoints)
	t time.Time

	// fileInfo holds the file info for the given file
	fileInfo os.FileInfo

	// IsRoot is a boolean flag stating for a root waypoint
	isRoot bool

	// Waypoints are inner children (subdirectories, files, etc)
	waypoints []Waypoint
}

type WaypointFiles []*WaypointFile

func (w *WaypointFile) Time() time.Time      { return w.fileInfo.ModTime() }
func (w *WaypointFile) Identifier() string   { return w.path }
func (w *WaypointFile) IsContainer() bool    { return w.fileInfo.IsDir() }
func (w *WaypointFile) Children() []Waypoint { return w.waypoints }

func NewWaypointFile(ctx context.Context, path string) (*WaypointFile, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	w := &WaypointFile{path: path, fileInfo: stat, t: stat.ModTime()}

	if stat.IsDir() {
		// Go deeper in the directory
		innerPaths, err := filepath.Glob(filepath.Join(w.path, "*"))
		if err != nil {
			return nil, err
		}

		for _, innerPath := range innerPaths {
			child, err := NewWaypointFile(ctx, innerPath)
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
