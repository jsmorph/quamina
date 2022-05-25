package main

import (
	"flag"
	"fmt"

	quamina "github.com/timbray/quamina/core"
	"github.com/timbray/quamina/internal/gen"
	"github.com/timbray/quamina/pruner"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// Warning: flag action below!

	var (
		mix  = gen.DefaultOpsMix
		spec = gen.DefaultCorpusSpec
		val  = gen.DefaultValue
		exec = gen.DefaultExec
	)

	flag.IntVar(&exec.Ops, "ops",
		exec.Ops, "ops per goroutine")
	flag.IntVar(&exec.Goroutines, "threads",
		exec.Goroutines, "number of goroutines")
	flag.IntVar(&spec.MatchingEvents, "matching-events",
		spec.MatchingEvents, "number of matching events")
	flag.IntVar(&spec.MatchingPatterns, "matching-patterns",
		spec.MatchingPatterns, "number of Matching patterns")
	flag.IntVar(&spec.OtherEvents, "other-events",
		spec.OtherEvents, "number of other events")
	flag.IntVar(&spec.OtherPatterns, "other-patterns",
		spec.OtherPatterns, "number of other patterns")
	flag.IntVar(&spec.PatternIds, "num-pattern-ids",
		spec.PatternIds, "number of pattern ids")

	var (
		core = flag.Bool("core", false, "use CoreMatcher instead of Pruner")
	)

	flag.Parse()

	if *core {
		exec.Matcher = quamina.NewCoreMatcher()
	} else {
		exec.Matcher = pruner.NewMatcher(nil)
	}

	exec.Mix = mix
	spec.V = val

	d, _, err := spec.Exec(exec)
	if err != nil {
		return err
	}
	fmt.Printf("elapsed %s\n", d)

	return nil
}
