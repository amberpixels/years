package years

import (
	"fmt"
	"slices"
)

// Voyager walks a waypoint tree in time order and finds waypoints by time.
type Voyager struct {
	root Waypoint

	// parser reads the time strings given to Navigate and Find.
	parser *Parser
}

// NewVoyager wraps root for traversal. The optional parser reads Navigate and
// Find queries; it defaults to NewParser().
func NewVoyager(root Waypoint, parserArg ...*Parser) *Voyager {
	v := &Voyager{root: root, parser: NewParser()}
	if len(parserArg) > 0 {
		v.parser = parserArg[0]
	} else {
		v.parser = NewParser()
	}

	return v
}

// TraverseDirection is a direction for traversing (e.g. past or future).
type TraverseDirection string

const (
	TraverseDirectionPast   TraverseDirection = "past"
	TraverseDirectionFuture TraverseDirection = "future"
)

// TraverseNodesMode specifies which type of nodes to traverse (e.g. leaves only or containers only).
type TraverseNodesMode string

const (
	TraverseLeavesOnly     TraverseNodesMode = "leaves_only"
	TraverseContainersOnly TraverseNodesMode = "containers_only"
	TraverseAllNodes       TraverseNodesMode = "all"
)

type traverseConfig struct {
	direction               TraverseDirection
	nodesMode               TraverseNodesMode
	includeNonCalendarNodes bool
}

// defaultTraverseConfig walks newest to oldest over every kind of node.
func defaultTraverseConfig() traverseConfig {
	return traverseConfig{
		direction: TraverseDirectionPast,
		nodesMode: TraverseAllNodes,
	}
}

// isTraversable checks if a given waypoint is traversable corresponding to config.
func (config *traverseConfig) isTraversable(waypoint Waypoint) bool {
	if waypoint.Time().IsZero() && !config.includeNonCalendarNodes {
		return false
	}

	if config.nodesMode == TraverseAllNodes {
		return true
	}

	okLeavesOnly := config.nodesMode == TraverseLeavesOnly && !waypoint.IsContainer()
	if okLeavesOnly {
		return true
	}
	okContainersOnly := config.nodesMode == TraverseContainersOnly && waypoint.IsContainer()
	return okContainersOnly
}

// TraverseOption defines functional options for the Traverse function.
type TraverseOption func(*traverseConfig)

// O_PAST returns a TraverseOption for traversing in Past direction.
func O_PAST() TraverseOption {
	return func(o *traverseConfig) { o.direction = TraverseDirectionPast }
}

// O_FUTURE returns a TraverseOption for traversing in Future direction.
func O_FUTURE() TraverseOption {
	return func(o *traverseConfig) { o.direction = TraverseDirectionFuture }
}

// O_LEAVES_ONLY returns a TraverseOption for traversing only leaf nodes.
func O_LEAVES_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseLeavesOnly }
}

// O_CONTAINERS_ONLY returns a TraverseOption for traversing only container nodes.
func O_CONTAINERS_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseContainersOnly }
}

// O_ALL returns a TraverseOption for traversing all nodes.
func O_ALL() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseAllNodes }
}

// O_NON_CALENDAR returns a TraverseOption for including non calendar nodes.
func O_NON_CALENDAR() TraverseOption {
	return func(o *traverseConfig) { o.includeNonCalendarNodes = true }
}

// Traverse calls cb for the root and every descendant in time order: newest
// first by default (O_PAST), oldest first with O_FUTURE. O_LEAVES_ONLY,
// O_CONTAINERS_ONLY and O_ALL (default) pick the nodes; waypoints with no time
// are skipped unless O_NON_CALENDAR is given.
func (v *Voyager) Traverse(cb func(w Waypoint), opts ...TraverseOption) error {
	config := defaultTraverseConfig()
	for _, opt := range opts {
		opt(&config)
	}

	var directionSign int
	switch config.direction {
	case TraverseDirectionPast:
		directionSign = -1
	case TraverseDirectionFuture:
		directionSign = 1
	default:
		panic("invalid traverse direction: " + config.direction)
	}

	sortFn := func(a, b Waypoint) int {
		if a.Time().Equal(b.Time()) {
			return directionSign
		}

		if a.Time().After(b.Time()) {
			return directionSign
		}

		return -directionSign
	}

	sorted := AllChildren(v.root)
	sorted = append(sorted, v.root)
	slices.SortFunc(sorted, sortFn)

	for _, sw := range sorted {
		if config.isTraversable(sw) {
			cb(sw)
		}
	}

	return nil
}

// Navigate returns the first waypoint, in traversal order, whose time equals
// the parsed to, or nil when none does: Navigate("yesterday") finds the
// waypoint at the start of yesterday.
func (v *Voyager) Navigate(to string) (Waypoint, error) {
	navigateTo, err := v.parser.Parse("", to)
	if err != nil {
		return nil, fmt.Errorf("could not parse time: %w", err)
	}

	var found Waypoint
	if err := v.Traverse(func(w Waypoint) {
		if found != nil {
			return
		}

		if w.Time().Equal(navigateTo) {
			found = w
			return
		}
	}); err != nil {
		return nil, fmt.Errorf("could not traverse: %w", err)
	}

	return found, nil
}

// Find returns every Waypoint whose time equals the parsed timeStr, e.g.
// Find("yesterday") returns all waypoints at the start of yesterday.
func (v *Voyager) Find(timeStr string) ([]Waypoint, error) {
	navigateTo, err := v.parser.Parse("", timeStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse time: %w", err)
	}

	found := make([]Waypoint, 0)
	if err := v.Traverse(func(w Waypoint) {
		if found != nil {
			return
		}

		if w.Time().Equal(navigateTo) {
			found = append(found, w)
			return
		}
	}); err != nil {
		return nil, fmt.Errorf("could not traverse: %w", err)
	}

	return found, nil
}
