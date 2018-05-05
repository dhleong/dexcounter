package main

import (
	"fmt"
	"os"

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

	counts, err := counter.Count(dep)
	if err != nil {
		panic(err)
	}

	total := counts.CalculateTotal()
	fmt.Printf("%s TOTALS:\n Methods: %d\n  Fields: %d\n",
		counts.Dependency,
		total.Methods,
		total.Fields,
	)

	fmt.Println("Tree:")
	dump(counts, "")
}

func dump(counts *model.Counts, indent string) {
	fmt.Println(indent, counts.Dependency, counts.OwnFields, counts.OwnMethods)
	for _, dep := range counts.Dependents {
		dump(dep, "  "+indent)
	}
}
