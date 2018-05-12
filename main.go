package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/docopt/docopt-go"

	dexcounter "github.com/dhleong/dexcounter/src"
	"github.com/dhleong/dexcounter/src/model"
)

func parseOptions() *model.Options {
	usage := `dexcounter: For counting methods. For Dex files.

Usage: dexcounter <dependency>

Options:
  --dx <pathToDx>  Path to dx executable; required if $ANDROID_HOME is not set
  -h, --help       Show this screen.
  --version        Show version.`
	args, _ := docopt.ParseArgs(
		usage,
		os.Args[1:],
		fmt.Sprintf("dexcounter version %s", dexcounter.Version),
	)

	options := model.Options{}
	args.Bind(&options)
	return &options
}

func main() {
	opts := parseOptions()

	counter, err := dexcounter.NewDexCounter(opts)
	if err != nil {
		panic(err)
	}

	dep, err := model.ParseDependency(opts.Dependency)
	if err != nil {
		panic(err)
	}

	clearLine := func() {
		fmt.Print("\r\x1b[2K")
	}

	fmt.Printf("Computing transitive dependencies...")
	var lock sync.Mutex
	totalTransitive := 0
	done := 0
	counts, err := counter.Count(
		dep,
		func(counts *model.Counts) {
			lock.Lock()
			totalTransitive = 1 + len(counts.Dependents)
			clearLine()
			fmt.Printf("Counting 0 / %d...", totalTransitive)
			lock.Unlock()
		},
		func(counts *model.Counts) {
			lock.Lock()
			done++
			clearLine()
			fmt.Printf("Counting %d / %d...", done, totalTransitive)
			lock.Unlock()
		},
	)
	if err != nil {
		panic(err)
	}

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
