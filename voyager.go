package years

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Waypoint struct {
	// Path of the waypoint
	// It's the request field for .prepare() to be called
	Path string

	// Name or the base of the path
	Name string

	// Time representing the start of its range (e.g. start of the day for daily waypoints)
	Time time.Time

	// Unit of waypoint representing the duration unit (day|month|year)
	Unit DateUnit

	// Waypoints are inner children (subdirectories, files, etc)
	Waypoints Waypoints
}

type Waypoints []*Waypoint

// prepare builds the waypoints tree
// layout is date layout e.g. "2006/Jan/02.txt", "2006/01-Jan/2006-01-02.txt", etc
func (w *Waypoint) prepare(layout string) error {
	stat, err := os.Stat(w.Path)
	if err != nil {
		return err
	}

	layoutParts := strings.Split(layout, string(os.PathSeparator))

	// currentLayout is the layout of the current file (under the cursor)
	// innerLayout is layout of inner objects (in case current is a directory)
	var currentLayout, innerLayout = layoutParts[0], layout

	w.Name = stat.Name()
	t, err := time.Parse(currentLayout, stat.Name())
	if err != nil { // todo: check if current step has to be a valid date
		log.Printf("Error parsing time from file %s: %v\n", w.Name, err)
	} else {
		w.Time = t

		layoutMeta := parseLayout(currentLayout)
		w.Unit = layoutMeta.MinimalUnit

		if len(layoutParts) > 0 {
			innerLayout = strings.Join(layoutParts[1:], string(os.PathSeparator))
		}
	}

	if !stat.IsDir() {
		return nil
	}

	// Go deeper in the directory:

	innerPaths, err := filepath.Glob(filepath.Join(w.Path, "*"))
	if err != nil {
		return err
	}

	for _, innerPath := range innerPaths {
		child := &Waypoint{Path: innerPath}
		if err := child.prepare(innerLayout); err != nil {
			log.Println("child failed: %w", err)
			continue
		}
		w.Waypoints = append(w.Waypoints, child)
	}

	return nil
}

type Voyager struct {
	root *Waypoint

	// layout is a complex path layout, that uses time.Layout's syntax for date/time
	// e.g. "2006/Jan/2006-01-02.txt"
	layout string
}

func NewVoyager(layout string) *Voyager {
	return &Voyager{layout: layout}
}

func (v *Voyager) Prepare(path string) error {
	v.root = &Waypoint{Path: path}
	return v.root.prepare(v.layout)
}

func (v *Voyager) WaypointsTree() *Waypoint { return v.root }
