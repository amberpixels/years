package years

import (
	"context"
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

// yearFromCtx/monthFromCtx are variables for ctx keys for storing
// year & month information about parent when iterating through calendar directory
// It's OK for a temporary solution
var (
	yearFromCtx  = "year"
	monthFromCtx = "month"
)

// prepare builds the waypoints tree
// layout is date layout e.g. "2006/Jan/02.txt", "2006/01-Jan/2006-01-02.txt", etc
func (w *Waypoint) prepare(ctx context.Context, layout string) error {
	stat, err := os.Stat(w.Path)
	if err != nil {
		return err
	}

	layoutParts := strings.Split(layout, string(os.PathSeparator))

	// currentLayout is the layout of the current file (under the cursor)
	// innerLayout is layout of inner objects (in case current is a directory)
	var currentLayout, innerLayout = layoutParts[0], layout

	// Parsing of current file's layout lets us know if we miss some parent date information
	// e.g. currentLayout is "01" (only month), then we miss parent's year
	//      or currentLayout is "02.txt" (only the day) then we miss both month and parent
	layoutMeta := parseLayout(currentLayout)
	var yearIsMissing, monthIsMissing bool = true, true
	for _, unitInLayout := range layoutMeta.Units {
		if unitInLayout == Year {
			yearIsMissing = false
		}
		if unitInLayout == Month {
			monthIsMissing = false
		}
	}

	w.Name = stat.Name()
	t, err := time.Parse(currentLayout, stat.Name())
	if err != nil { // todo: check if current step has to be a valid date
		log.Printf("Error parsing time from file %s: %v\n", w.Name, err)
	} else {
		w.Time = t
		w.Unit = layoutMeta.MinimalUnit

		// TODO: reconsider. It's a weak solution for now
		if w.Unit < Year {
			if yearIsMissing {
				if yearsFromParent, ok := ctx.Value(yearFromCtx).(int); ok {
					New(&w.Time).SetYear(yearsFromParent)
				}
			}
		}
		if w.Unit < Month {
			if monthIsMissing {
				if monthFromParent, ok := ctx.Value(monthFromCtx).(time.Month); ok {
					New(&w.Time).SetMonth(monthFromParent)
				}
			}
		}

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

		childCtx := ctx
		if w.Unit <= Year {
			childCtx = context.WithValue(childCtx, yearFromCtx, w.Time.Year())
		}
		if w.Unit <= Month {
			childCtx = context.WithValue(childCtx, monthFromCtx, w.Time.Month())
		}

		if err := child.prepare(childCtx, innerLayout); err != nil {
			log.Println("child failed: %w", err)
			continue
		}
		// Inserting child into the list of waypoints, but respecting the order
		// To achieve this, we use simple append first time for first child
		// And then we insert on the `index` position first time it met time earlier than new child
		// Note: it's not the most optimal solution for inserting at position `index`
		//       In future this is a place to be optimized
		var inserted bool
		for index, existedChild := range w.Waypoints {
			if existedChild.Time.Before(child.Time) {
				continue
			}
			w.Waypoints = append(w.Waypoints[:index+1], w.Waypoints[index:]...)
			w.Waypoints[index] = child
			inserted = true
			break
		}

		// if not inserted, then it's the newest, so insert the last
		if !inserted {
			w.Waypoints = append(w.Waypoints, child)
		}
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
	return v.root.prepare(context.Background(), v.layout)
}

func (v *Voyager) WaypointsTree() *Waypoint { return v.root }
