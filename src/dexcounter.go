package dexcounter

import (
	"github.com/dhleong/dexcounter/src/counters"
	"github.com/dhleong/dexcounter/src/model"
)

// Version is the current version of the app
const Version = "1.1.0"

// NewDexCounter creates a DexCounter that combines
// all the strategies as well as possible
func NewDexCounter(opts *model.Options) (model.DexCounter, error) {
	gradleCounter, err := counters.NewGradleDexCounter()
	if err != nil {
		return nil, err
	}
	return counters.NewDxDexCounter(opts, gradleCounter)
}
