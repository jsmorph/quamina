package gen

import (
	"runtime"

	"github.com/timbray/quamina/pruner"
)

var (
	DefaultCorpusSpec = &CorpusSpec{
		// Need V: Value
		Trimmer: &Trimmer{
			Map:   0.5,
			Array: 0.5,
		},
		MatchingEvents:   1000,
		MatchingPatterns: 1000,
		OtherEvents:      1000,
		OtherPatterns:    1000,
		PatternIds:       400,
	}

	DefaultOpsMix = &OpsMix{
		Adds:    10,
		Matches: 20,
		Deletes: 5,
	}

	DefaultExec = &Exec{
		// Need Mix
		Matcher:    pruner.NewMatcher(nil),
		Goroutines: runtime.NumCPU(),
		Filename:   "",
		Ops:        1000,
	}

	DefaultValue = &Value{
		Map:    0.5,
		Array:  0.5,
		Int:    0.4,
		String: 0.6,
		Strings: String{
			Length: Int{
				Min: 5,
				Max: 30,
			},
		},
		Arrays: Array{
			Length: Int{
				Min: 1,
				Max: 5,
			},
		},
		Ints: Int{
			Min: -100,
			Max: 100,
		},
		Maps: Map{
			NumProperties: Int{
				Min: 2,
				Max: 5,
			},
			Properties: String{
				Length: Int{
					Min: 3,
					Max: 20,
				},
			},
		},
		Decays: Decays{
			Map:   0.5,
			Array: 0.8,
		},
	}
)

func Defaults() (*CorpusSpec, *Exec) {
	c := DefaultCorpusSpec.Copy()
	c.V = DefaultValue.Copy()
	e := DefaultExec.Copy()
	e.Mix = DefaultOpsMix.Copy()
	return c, e
}
