package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	corpusgen "github.com/jsmorph/corpusgen/gen"

	"github.com/timbray/quamina"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	rand.Seed(time.Now().UnixNano())

	var (
		numEvents        = flag.Int("events", 10000, "number of events")
		patternsPerEvent = flag.Int("patterns-per-event", 3, "patterns per event")
		matchingPatterns = flag.Int("matching-patterns", 1000, "matching patterns")
		numPatterns      = flag.Int("patterns", 2000, "number of patterns")
		step             = flag.Int("step", 200, "patterns count step")
		goroutines       = flag.Int("goroutines", 10, "number of goroutines for 'concurrent' mode")
		mode             = flag.String("mode", "events", "'patterns', 'events', 'showgen', or 'concurrent'")

		numPropsMin = flag.Int("min-props", 5, "starting minimum number of properties for events")
		numPropsMax = flag.Int("max-props", 5, "starting maximum number of properties for events")

		events   [][]byte
		patterns []string
	)

	flag.Parse()

	valid := func(p string) bool {
		q, _ := quamina.New()
		return nil == q.AddPattern(true, p)
	}

	cfg := struct {
		Value   *corpusgen.Value
		Trimmer *corpusgen.Trimmer
	}{
		Value:   corpusgen.DefaultValue,
		Trimmer: corpusgen.DefaultTrimmer,
	}

	// Override those defaults.
	cfg.Value.Map = 0.5
	cfg.Value.Array = 0.2
	cfg.Value.Int = 0.1
	cfg.Value.String = 0.2
	cfg.Value.Maps.NumProperties = corpusgen.Int{
		Min: *numPropsMin,
		Max: *numPropsMax,
	}

	then := time.Now()
	for len(events) < *numEvents {
		event, err := cfg.Value.GenerateEvent(true)
		if err != nil {
			return err
		}
		events = append(events, event)
		for j := 0; j < *patternsPerEvent && len(patterns) < *matchingPatterns; j++ {
			pattern, err := cfg.Trimmer.DerivePattern(event)
			if err != nil {
				return err
			}
			if !valid(pattern) {
				continue
			}
			patterns = append(patterns, pattern)
		}
	}

	for len(patterns) < *numPatterns {
		event, err := cfg.Value.GenerateEvent(true)
		if err != nil {
			return err
		}
		pattern, err := cfg.Trimmer.DerivePattern(event)
		if err != nil {
			return err
		}
		if !valid(pattern) {
			continue
		}
		patterns = append(patterns, pattern)
	}

	rand.Shuffle(len(patterns),
		func(i, j int) { patterns[i], patterns[j] = patterns[j], patterns[i] })

	log.Printf("generated %d patterns and %d events in %s",
		len(patterns), len(events), time.Now().Sub(then))

	q, err := quamina.New()
	if err != nil {
		return err
	}

	switch *mode {
	case "showgen":
		// Just print the generated patterns and events.

		for _, p := range patterns {
			fmt.Printf("pattern\t%s\n", p)
		}
		for _, e := range events {
			fmt.Printf("event\t%s\n", e)
		}

	case "patterns":
		// AddPatterns only.

		fmt.Printf("patterns,msPerPattern,msPerPatternTail\n")
		then := time.Now()
		for n := 0; n < len(patterns); n += *step {
			t0 := time.Now()
			for i, pattern := range patterns[n : n+*step] {
				if err := q.AddPattern(i, pattern); err != nil {
					return err
				}
			}
			var (
				elapsed        = time.Now().Sub(then)
				rateMillis     = 1000 * elapsed.Seconds() / float64(n+*step)
				tailRateMillis = 1000 * time.Now().Sub(t0).Seconds() / float64(*step)
			)

			fmt.Printf("%d,%f,%f\n", n+*step, rateMillis, tailRateMillis)
		}

	case "events":
		// AddPatterns then time matches.

		fmt.Printf("patterns,events,msPerEvent,msPerEventTail,matched\n")
		var totalElapsed time.Duration
		var totalEvents int

		for n := 0; n < len(patterns); n += *step {
			for i, pattern := range patterns[n : n+*step] {
				if err := q.AddPattern(i, pattern); err != nil {
					return err
				}
			}

			matched := 0
			then := time.Now()
			for _, event := range events {
				xs, err := q.MatchesForEvent(event)
				if err != nil {
					return err
				}
				matched += len(xs)
			}
			totalEvents += len(events)
			elapsed := time.Now().Sub(then)
			totalElapsed += elapsed
			rateMillis := 1000 * totalElapsed.Seconds() / float64(totalEvents)
			tailRateMillis := 1000 * elapsed.Seconds() / float64(len(events))

			fmt.Printf("%d,%d,%f,%f,%d\n",
				n+*step, len(events), rateMillis, tailRateMillis, matched)
		}
	case "concurrent":
		fmt.Printf("goroutines,patterns,copySecs,events,msPerEvent,elapsedSecs,matched\n")
		then = time.Now()
		for i, pattern := range patterns {
			if err := q.AddPattern(i, pattern); err != nil {
				return err
			}
		}
		elapsed := time.Now().Sub(then)
		log.Printf("added %d patterns in %s (%f ms per pattern)",
			len(patterns), elapsed, 1000*elapsed.Seconds()/float64(len(patterns)))

		wg := &sync.WaitGroup{}
		for i := 0; i < *goroutines; i++ {
			wg.Add(1)
			then = time.Now()
			q0 := q.Copy()

			go func(i int, q *quamina.Quamina, copyElapsed time.Duration) {
				defer wg.Done()

				matched := 0
				then = time.Now()
				for _, event := range events {
					xs, err := q.MatchesForEvent(event)
					if err != nil {
						log.Printf("WARNING: match error %s on %s", err, event)
					}
					matched += len(xs)
				}
				elapsed := time.Now().Sub(then)
				rateMillis := 1000 * elapsed.Seconds() / float64(len(events))

				fmt.Printf("%d,%d,%f,%d,%f,%f,%d\n",
					*goroutines,
					len(patterns), copyElapsed.Seconds(),
					len(events), rateMillis, elapsed.Seconds(), matched)
			}(i, q0, time.Now().Sub(then))
		}
		wg.Wait()
	default:
		log.Fatalf("need -mode: events|patterns|showgen|concurrent")
	}

	return nil
}
