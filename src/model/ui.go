package model

// UI abstracts how the data is presented to the user.
// Its methods are called on background goroutines.
type UI interface {

	// OnStartResolve is called when the process to
	// resolve the transient dependencies has started
	OnStartResolve()

	// OnDependenciesResolved is called when all the
	// transitive dependencies have been resolved, but
	// their method counts may not yet have been found
	OnDependenciesResolved(counts *Counts)

	// OnDependencyCounted is called when the provided
	// dependency's methods have been counted.
	OnDependencyCounted(counts *Counts)

	// OnDone is called with the final results of the
	// Count process
	OnDone(counts *Counts)

	// OnError .
	OnError(error)
}
