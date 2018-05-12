package ui

import (
	"fmt"
	"sync"

	"github.com/dhleong/dexcounter/src/model"
)

type dumbUI struct {
	sync.Mutex

	totalTransitive int
	done            int
}

// NewDumbUI creates a dumb, CLI-based UI that may not
// be platform-independent
func NewDumbUI() model.UI {
	return &dumbUI{}
}

func (ui *dumbUI) OnStartResolve() {
	fmt.Printf("Computing transitive dependencies...")
}

func (ui *dumbUI) OnDependenciesResolved(counts *model.Counts) {
	total := 1 + len(counts.Dependents)

	ui.Lock()
	ui.totalTransitive = total
	clearLine()
	fmt.Printf("Counting 0 / %d...", total)
	ui.Unlock()
}

func (ui *dumbUI) OnDependencyCounted(count *model.Counts) {
	ui.Lock()
	ui.done++
	clearLine()
	fmt.Printf("Counting %d / %d...", ui.done, ui.totalTransitive)
	ui.Unlock()
}

func (ui *dumbUI) OnDone(counts *model.Counts) {

	clearLine()
	total := counts.CalculateTotal()
	fmt.Printf("%s TOTALS:\n Methods: %d\n  Fields: %d\n",
		counts.Dependency,
		total.Methods,
		total.Fields,
	)

	totalNameWidth := len(counts.Dependency.String())
	for _, dep := range counts.Dependents {
		width := len(dep.String()) + 2
		if width > totalNameWidth {
			totalNameWidth = width
		}
	}

	// extra padding
	totalNameWidth += 2

	format := fmt.Sprintf(
		"\n%%-%dsMethods  Fields\n",
		totalNameWidth,
	)
	fmt.Printf(format, "Dependency")
	dump(counts, totalNameWidth, "")
}

func (ui *dumbUI) OnError(err error) {
	panic(err)
}

func clearLine() {
	fmt.Print("\r\x1b[2K")
}

func dump(counts *model.Counts, totalNameWidth int, indent string) {
	format := fmt.Sprintf(
		"%%s%%-%ds  %%5d   %%5d\n",
		totalNameWidth-len(indent),
	)
	fmt.Printf(
		format,
		indent,
		counts.Dependency,
		counts.OwnMethods,
		counts.OwnFields,
	)
	for _, dep := range counts.Dependents {
		dump(dep, totalNameWidth, "  "+indent)
	}
}
