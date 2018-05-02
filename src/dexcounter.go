package dexcounter

import (
	"github.com/dhleong/dexcounter/src/counters"
	"github.com/dhleong/dexcounter/src/model"
)

// NewDexCounter creates a DexCounter that combines
// all the strategies as well as possible
func NewDexCounter() (model.DexCounter, error) {
	gradleCounter, err := counters.NewGradleDexCounter()
	if err != nil {
		return nil, err
	}
	return counters.NewDxDexCounter(gradleCounter)
}
