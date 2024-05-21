package years

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WaypointFile is a Waypoint implementation for files/directories
type WaypointFile struct {
	// Path of the waypoint
	// It's the request field for .prepare() to be called
	path string

	// stat holds the file info for the given file
	stat *os.File

	// timeInput is the full input that is using for time.Parse()
	// Usually it's considered to be file/dir name + names of parents (if necessary)
	timeInput string

	// ownLayout is layout of the file/dir itself
	// It can be partial as probably parent information is required for full layout
	// e.g. "Jan" (not knowing the year here)
	ownLayout string

	// layout is a full layout (knowing required parent's layout information)
	// e.g. "2006/Jan"
	layout string

	// Unit of waypoint representing the duration unit (day|month|year)
	unit DateUnit

	// IsDir is a boolean flag stating for directory waypoints
	isDir bool

	// IsRoot is a boolean flag stating for a root waypoint
	isRoot bool

	// Time representing the start of its range (e.g. start of the day for daily waypoints)
	t time.Time

	// Waypoints are inner children (subdirectories, files, etc)
	waypoints WaypointFiles
}

type WaypointFiles []*WaypointFile

func (w *WaypointFile) Time() time.Time    { return w.t }
func (w *WaypointFile) Identifier() string { return w.path }
func (w *WaypointFile) IsContainer() bool  { return w.isDir }
func (w *WaypointFile) Children() ([]Waypoint, error) {
	children := make([]Waypoint, len(w.waypoints))
	for i, waypoint := range w.waypoints {
		children[i] = waypoint
	}
	return children, nil
}

func WithCtxWaypointFileGlobalLayout(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, "global_layout", v)
}

func globalLayoutFromCtx(ctx context.Context) string {
	return ctx.Value("global_layout").(string)
}

func withCtxWaypointFileParent(ctx context.Context, v *WaypointFile) context.Context {
	return context.WithValue(ctx, "parent", v)
}
func parentFromCtx(ctx context.Context) *WaypointFile {
	p, _ := ctx.Value("parent").(*WaypointFile)
	return p
}

func NewWaypointFile(ctx context.Context, path string) (*WaypointFile, error) {
	globalLayout := globalLayoutFromCtx(ctx)
	if globalLayout == "" {
		panic("global_layout is required")
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	w := &WaypointFile{path: path, isDir: stat.IsDir()}

	parent := parentFromCtx(ctx)
	if parent == nil {
		w.isRoot = true
	}

	var ownLayout, layout string
	globalLayoutParts := strings.Split(globalLayout, string(os.PathSeparator))

	if w.isRoot {
		ownLayout = globalLayoutParts[0]
		layout = ownLayout
	} else {
		ownLayout = strings.TrimPrefix(globalLayout, parent.layout+"/")
		if w.isDir {
			ownLayout = strings.Split(ownLayout, string(os.PathSeparator))[0]
		}
		layout = parent.layout + string(os.PathSeparator) + ownLayout
	}

	layout = strings.TrimPrefix(layout, string(os.PathSeparator))
	ownLayout = strings.TrimPrefix(ownLayout, string(os.PathSeparator))

	w.timeInput = stat.Name()
	if parent != nil && parent.timeInput != "" {
		w.timeInput = parent.timeInput + string(os.PathSeparator) + w.timeInput
	}
	w.layout = layout
	w.ownLayout = ownLayout

	// TODO: fix this
	if stat.Name() == "calendar1" || stat.Name() == "calendar2" {
		w.layout = ""
		w.timeInput = ""
		w.ownLayout = ""
	}

	if w.timeInput != "" {
		t, err := time.Parse(layout, w.timeInput)
		if err != nil {
			log.Printf("Error parsing time from file %s: %v\n", w.timeInput, err)
		} else {
			w.t = t
			lm := parseLayout(layout)
			w.unit = lm.MinimalUnit
		}
	}

	if stat.IsDir() {

		// Go deeper in the directory
		innerPaths, err := filepath.Glob(filepath.Join(w.path, "*"))
		if err != nil {
			return nil, err
		}

		for _, innerPath := range innerPaths {
			child, err := NewWaypointFile(withCtxWaypointFileParent(ctx, w), innerPath)
			if err != nil {
				log.Println("child failed: %w", err)
				continue
			}

			// By default, let's sort nodes in Past->Future order
			var inserted bool
			for i, existed := range w.waypoints {
				if existed.t.After(child.t) {
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
