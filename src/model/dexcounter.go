package model

// DepsComputed is a callback when all the transitive
// dependencies have been resolved, but their method
// counts may not yet have been found
type DepsComputed = func(*Counts)

// DepCounted is a callback when a dependency's methods
// have been counted
type DepCounted = func(*Counts)

// DexCounter is the interface for checking dex method/field counts
type DexCounter interface {
	Count(
		dependency Dependency,
		onDepsComputed DepsComputed,
		onDepCounted DepCounted,
	) (*Counts, error)
}
