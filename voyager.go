package years

import (
	"fmt"
)

type Voyager struct {
	root Waypoint
}

func NewVoyager(start Waypoint) *Voyager {
	return &Voyager{root: start}
}

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
	TraverseLeavesOnly     TraverseNodes = "leaves_only"
	TraverseContainersOnly TraverseNodes = "containers_only"
	TraverseAllNodes       TraverseNodes = "all"
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

// isTraversable checks if a given waypoint is traversable corresponding to config
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
	if okContainersOnly {
		return true
	}

	return false
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

// O_LEAVES_ONLY returns a TraverseOption for traversing only leaf nodes
func O_LEAVES_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseLeavesOnly }
}

// O_CONTAINERS_ONLY returns a TraverseOption for traversing only container nodes
func O_CONTAINERS_ONLY() TraverseOption {
	return func(o *traverseConfig) { o.nodesMode = TraverseContainersOnly }
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
func (v *Voyager) Traverse(cb func(w Waypoint), opts ...TraverseOption) error {
	config := defaultTraverseConfig()
	for _, opt := range opts {
		opt(&config)
	}

	switch config.direction {
	case TraverseDirectionPast:
		return v.traversePast(v.root, cb, &config)
	case TraverseDirectionFuture:
		return v.traverseFuture(v.root, cb, &config)
	default:
		panic("invalid traverse direction: " + config.direction)
	}
}

func (v *Voyager) traversePast(waypoint Waypoint, cb func(w Waypoint), config *traverseConfig) error {
	if waypoint == nil {
		return nil
	}

	children := waypoint.Children()

	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		if err := v.traversePast(child, cb, config); err != nil {
			return fmt.Errorf("failed to traversePast a child: %w", err)
		}
	}

	if config.isTraversable(waypoint) {
		cb(waypoint)
	}

	return nil
}

func (v *Voyager) traverseFuture(waypoint Waypoint, cb func(w Waypoint), config *traverseConfig) error {
	if waypoint == nil {
		return nil
	}

	if config.isTraversable(waypoint) {
		cb(waypoint)
	}

	children := waypoint.Children()

	for _, child := range children {
		if err := v.traverseFuture(child, cb, config); err != nil {
			return fmt.Errorf("failed to traverseFuture a child: %w", err)
		}
	}

	return nil
}

func (v *Voyager) Navigate(to string) (Waypoint, error) {
	navigateTo, _ := NewParser(
		AcceptAliases(),
		AcceptUnix(),
		// todo: supported layouts should be known from voyager
		WithLayouts("2006-01-02"),
	).ParseTime(to)

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
