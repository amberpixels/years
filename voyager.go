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

	// Layout on which time was parsed
	Layout string

	// Time representing the start of its range (e.g. start of the day for daily waypoints)
	Time time.Time

	// Unit of waypoint representing the duration unit (day|month|year)
	Unit DateUnit

	// IsDir is a boolean flag stating for directory waypoints
	IsDir bool

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

	if stat.IsDir() {
		w.IsDir = true
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
	w.Layout = currentLayout
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
					Wrap(&w.Time).SetYear(yearsFromParent)
				}
			}
		}
		if w.Unit < Month {
			if monthIsMissing {
				if monthFromParent, ok := ctx.Value(monthFromCtx).(time.Month); ok {
					Wrap(&w.Time).SetMonth(monthFromParent)
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

// Traversing means walking through voyager's prepared tree

// TraverseDirection TODO nicer enum-ish
type TraverseDirection string

const (
	TraverseDirectionPast   TraverseDirection = "past"
	TraverseDirectionFuture TraverseDirection = "future"
)

// TraverseNodes TODO nicer enum-ish
type TraverseNodes string

const (
	TraverseFilesOnly TraverseNodes = "files_only"
	TraverseDirsOnly  TraverseNodes = "dirs_only"
	TraverseAllNodes  TraverseNodes = "all"
)

type traverseConfig struct {
	direction               TraverseDirection
	nodesMode               TraverseNodes
	includeNonCalendarNodes bool
}

// defaultTraverseConfig is Future->Past + all type of nodes
func defaultTraverseConfig() traverseConfig {
	return traverseConfig{
		direction: TraverseDirectionPast,
		nodesMode: TraverseAllNodes,
	}
}

// isTraversable checks if a given waypoint is traversable regard to config
func (config *traverseConfig) isTraversable(waypoint *Waypoint) bool {
	if waypoint.Time.IsZero() && !config.includeNonCalendarNodes {
		return false
	}

	if config.nodesMode == TraverseAllNodes {
		return true
	}

	okFileOnly := config.nodesMode == TraverseFilesOnly && !waypoint.IsDir
	okDirOnly := config.nodesMode == TraverseDirsOnly && waypoint.IsDir
	return okFileOnly || okDirOnly
}

// TraverseOption defines functional options for the Traverse function
type TraverseOption func(*traverseConfig)

// O_PAST returns a TraverseOption for traversing in Past direction
func O_PAST() TraverseOption {
	return func(o *traverseConfig) { o.direction = TraverseDirectionPast }
}

// O_FUTURE returns a TraverseOption for traversing in Future direction
func O_FUTURE() TraverseOption {
	return func(o *traverseConfig) { o.direction = TraverseDirectionFuture }
}

// O_FILES_ONLY returns a TraverseOption for traversing only file nodes
func O_FILES_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseFilesOnly }
}

// O_DIRS_ONLY returns a TraverseOption for traversing only dir nodes
func O_DIRS_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseDirsOnly }
}

// O_ALL returns a TraverseOption for traversing all nodes
func O_ALL() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseAllNodes }
}

// O_NON_CALENDAR returns a TraverseOption for including non calendar nodes
func O_NON_CALENDAR() TraverseOption {
	return func(o *traverseConfig) { o.includeNonCalendarNodes = true }
}

// Traverse traverses the built voyager tree in the given direction
func (v *Voyager) Traverse(cb func(w *Waypoint), opts ...TraverseOption) {
	config := defaultTraverseConfig()
	for _, opt := range opts {
		opt(&config)
	}

	switch config.direction {
	case TraverseDirectionPast:
		v.traversePast(v.root, cb, &config)
	case TraverseDirectionFuture:
		v.traverseFuture(v.root, cb, &config)
	default:
		panic("invalid traverse direction: " + config.direction)
	}
}

func (v *Voyager) traversePast(waypoint *Waypoint, cb func(w *Waypoint), config *traverseConfig) {
	if waypoint == nil {
		return
	}

	for i := len(waypoint.Waypoints) - 1; i >= 0; i-- {
		child := waypoint.Waypoints[i]
		v.traversePast(child, cb, config)
	}

	if config.isTraversable(waypoint) {
		cb(waypoint)
	}
}

func (v *Voyager) traverseFuture(waypoint *Waypoint, cb func(w *Waypoint), config *traverseConfig) {
	if waypoint == nil {
		return
	}

	if config.isTraversable(waypoint) {
		cb(waypoint)
	}

	for _, child := range waypoint.Waypoints {
		v.traverseFuture(child, cb, config)
	}
}

func (v *Voyager) Navigate(to string) *Waypoint {
	// todo: supported layouts should be known from voyager
	navigateTo, _ := NewParser(WithLayouts("2006-01-02")).ParseTime(to)
	var found *Waypoint
	v.Traverse(func(w *Waypoint) {
		if found != nil {
			return
		}

		if w.Time.Equal(navigateTo) {
			found = w
			return
		}
	})

	return found
}
