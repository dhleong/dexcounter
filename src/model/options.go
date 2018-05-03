package model

// Options provided by CLI flags, etc.
type Options struct {
	Dependency string

	DxPath string `docopt:"dx"`
}
