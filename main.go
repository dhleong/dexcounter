package main

import (
	"fmt"

	"github.com/docopt/docopt-go"

	dexcounter "github.com/dhleong/dexcounter/src"
	"github.com/dhleong/dexcounter/src/model"
)

func main() {
	opts, _ := docopt.ParseDoc(
		`dexcounter

Usage: dexcounter [<dependency>]
		`,
	)

	counter, err := dexcounter.NewDexCounter()
	if err != nil {
		panic(err)
	}

	depString := opts["<dependency>"]
	if depString == nil {
		// for dev purposes this arg is optional,
		// but it should not be...
		depString = "com.squareup.okhttp3:okhttp:3.10.0"
	}

	dep, err := model.ParseDependency(depString.(string))
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
