package years

import (
	"fmt"
	"slices"
)

// Voyager is a wrapper for a waypoint that allows for traversing through it.
type Voyager struct {
	root Waypoint

	// parser is used for parsing time when needed (e.g. when navigating).
	parser *Parser
}

func NewVoyager(root Waypoint, parserArg ...*Parser) *Voyager {
	v := &Voyager{root: root, parser: NewParser()}
	if len(parserArg) > 0 {
		v.parser = parserArg[0]
	} else {
		v.parser = NewParser()
	}

	return v
}

// Traversing means walking through voyager's prepared tree.

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

// defaultTraverseConfig is Future->Past + all type of nodes.
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
//
//nolint:revive,stylecheck,staticcheck // ok
func O_PAST() TraverseOption {
	return func(o *traverseConfig) { o.direction = TraverseDirectionPast }
}

// O_FUTURE returns a TraverseOption for traversing in Future direction.
//
//nolint:revive,stylecheck,staticcheck // ok
func O_FUTURE() TraverseOption {
	return func(o *traverseConfig) { o.direction = TraverseDirectionFuture }
}

// O_LEAVES_ONLY returns a TraverseOption for traversing only leaf nodes.
//
//nolint:revive,stylecheck,staticcheck // ok
func O_LEAVES_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseLeavesOnly }
}

// O_CONTAINERS_ONLY returns a TraverseOption for traversing only container nodes.
//
//nolint:revive,stylecheck,staticcheck // ok
func O_CONTAINERS_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseContainersOnly }
}

// O_ALL returns a TraverseOption for traversing all nodes.
//
//nolint:revive,stylecheck,staticcheck // ok
func O_ALL() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseAllNodes }
}

// O_NON_CALENDAR returns a TraverseOption for including non calendar nodes.
//
//nolint:revive,stylecheck,staticcheck // ok
func O_NON_CALENDAR() TraverseOption {
	return func(o *traverseConfig) { o.includeNonCalendarNodes = true }
}

// Traverse traverses through a given waypoint (all its children recursively).
func (v *Voyager) Traverse(cb func(w Waypoint), opts ...TraverseOption) error {
	config := defaultTraverseConfig()
	for _, opt := range opts {
		opt(&config)
	}

	// directionSign will be used in sorting func
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

// Navigate returns the first found Waypoint that matches given time (as a string).
// E.g. Navigate("yesterday") returns waypoint corresponding to the yesterday's date.
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

// Find returns the all found Waypoints that match given time (as a string)
// e.g. Find("yesterday") returns all waypoints whose time is in the "yesterday" range.
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
