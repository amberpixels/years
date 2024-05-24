package years

import (
	"context"
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

	// ownLayout is layout of the file/dir itself
	// It can be partial as probably parent information is required for full layout
	// e.g. "Jan" (not knowing the year here)
	ownLayout string

	// layout is a full layout (knowing required parent's layout information)
	// e.g. "2006/Jan"
	layout string

	// Unit of waypoint representing the duration unit (day|month|year)
	unit DateUnit
}

func (w *TimeNamedWaypointFile) setNonCalendar() {
	w.layout = ""
	w.timeInput = ""
	w.ownLayout = ""
	w.timeInput = ""
	w.t = time.Time{}
}

type TimeNamedWaypointFiles []*TimeNamedWaypointFile

func (w *TimeNamedWaypointFile) Time() time.Time { return w.t }

func WithCtxWaypointFileGlobalLayout(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, "global_layout", v)
}

func globalLayoutFromCtx(ctx context.Context) string {
	return ctx.Value("global_layout").(string)
}

func withCtxWaypointFileParent(ctx context.Context, v *TimeNamedWaypointFile) context.Context {
	return context.WithValue(ctx, "parent", v)
}
func parentFromCtx(ctx context.Context) *TimeNamedWaypointFile {
	p, _ := ctx.Value("parent").(*TimeNamedWaypointFile)
	return p
}

func NewTimeNamedWaypointFile(ctx context.Context, path string) (*TimeNamedWaypointFile, error) {
	globalLayout := globalLayoutFromCtx(ctx)
	if globalLayout == "" {
		panic("global_layout is required")
	}

	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	w0 := &WaypointFile{path: path, fileInfo: stat, t: stat.ModTime()}
	w := &TimeNamedWaypointFile{WaypointFile: w0}

	var ownLayout, layout string
	globalLayoutParts := strings.Split(globalLayout, string(os.PathSeparator))

	parent := parentFromCtx(ctx)
	if parent != nil {
		ownLayout = strings.TrimPrefix(globalLayout, parent.layout+"/")
		if w.fileInfo.IsDir() {
			ownLayout = strings.Split(ownLayout, string(os.PathSeparator))[0]
		}
		layout = parent.layout + string(os.PathSeparator) + ownLayout
	} else {
		ownLayout = globalLayoutParts[0]
		layout = ownLayout
	}

	layout = strings.TrimPrefix(layout, string(os.PathSeparator))
	ownLayout = strings.TrimPrefix(ownLayout, string(os.PathSeparator))

	w.timeInput = w.fileInfo.Name()
	if parent != nil && parent.timeInput != "" {
		w.timeInput = parent.timeInput + string(os.PathSeparator) + w.timeInput
	}
	w.layout = layout
	w.ownLayout = ownLayout

	lm := parseLayout(layout)
	if lm == nil {
		w.setNonCalendar()
	} else {
		t, err := time.Parse(layout, w.timeInput)
		if err == nil {
			w.t = t
			w.unit = lm.MinimalUnit
		} else {
			w.setNonCalendar()
		}
	}

	if w.fileInfo.IsDir() {

		// Go deeper in the directory
		innerPaths, err := filepath.Glob(filepath.Join(w.path, "*"))
		if err != nil {
			return nil, err
		}

		for _, innerPath := range innerPaths {
			child, err := NewTimeNamedWaypointFile(withCtxWaypointFileParent(ctx, w), innerPath)
			if err != nil {
				log.Println("child failed: %w", err)
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
