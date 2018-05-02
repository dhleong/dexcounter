package model

import (
	"errors"
	"fmt"
	"strings"
)

// Dependency .
type Dependency struct {
	Group    string
	Artifact string
	Version  string
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s:%s:%s", d.Group, d.Artifact, d.Version)
}

// ParseDependency accepts a dependency in `artifact:group:version`
// format and returns the parsed Dependency
func ParseDependency(dependencyString string) (Dependency, error) {
	parts := strings.Split(dependencyString, ":")
	if len(parts) != 3 {
		return Dependency{}, errors.New("Invalid dependency format")
	}

	return Dependency{
		Group:    parts[0],
		Artifact: parts[1],
		Version:  parts[2],
	}, nil
}

// Counts .
type Counts struct {
	Dependency
	Path       string
	OwnMethods int
	OwnFields  int

	Dependents []*Counts
}

// TotalCounts .
type TotalCounts struct {
	Methods int
	Fields  int
}

// Flatten returns a map of Dependency to its *Counts,
// to deduplicate on Dependency
func (c *Counts) Flatten() map[Dependency]*Counts {
	result := map[Dependency]*Counts{
		c.Dependency: c,
	}

	for _, dep := range c.Dependents {
		depMap := dep.Flatten()
		for k, v := range depMap {
			result[k] = v
		}
	}

	return result
}

// CalculateTotal .
func (c *Counts) CalculateTotal() TotalCounts {
	flattened := c.Flatten()
	methods := 0
	fields := 0

	for _, counts := range flattened {
		methods += counts.OwnMethods
		fields += counts.OwnFields
	}

	return TotalCounts{
		methods,
		fields,
	}
}

func (c *Counts) totalWithMap(seen map[string]bool) TotalCounts {
	methods := c.OwnMethods
	fields := c.OwnFields

	for _, d := range c.Dependents {
		childTotals := d.CalculateTotal()
		methods += childTotals.Methods
		fields += childTotals.Fields
	}

	return TotalCounts{
		methods,
		fields,
	}
}
