package model

// DexCounter is the interface for checking dex method/field counts
type DexCounter interface {
	Count(dependency Dependency) (*Counts, error)
}
