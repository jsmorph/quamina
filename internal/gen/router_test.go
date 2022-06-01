package gen

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/timbray/quamina/core"
)

// TestRouter runs some concurrent consumers and producers who deal in
// events and patterns randomly selected from a generated corpus.
func TestRouter(t *testing.T) {
	var (
		numConsumers  = 1000
		numForwarders = 4
		numEvents     = 10000

		r           = NewRouter()
		wg          = &sync.WaitGroup{}
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		forwarded   uint64

		pause = func() {
			ms := rand.Intn(50)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	)
	defer cancel()

	spec := DefaultCorpusSpec.Copy()
	spec.V = DefaultValue.Copy()
	corpus, err := spec.Gen()
	if err != nil {
		t.Fatal(err)
	}

	// Monitor

	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			case <-time.NewTicker(time.Second).C:
				log.Printf("forwarded %d", atomic.LoadUint64(&forwarded))
			}
		}
	}()

	// Forwarders
	events := append(corpus.MatchingEvents, corpus.OtherEvents...)
	for i := 0; i < numForwarders; i++ {
		go func(i int) {
			f := core.NewFJ(r.m.Matcher)

			for j := 0; j < numEvents/numForwarders; j++ {
				pause()
				for {
					event := events[rand.Intn(len(events))]
					if err := r.Route(ctx, f, event); err == nil {
						// log.Printf("publishing %s", event)
						break
					} else {
						log.Println(err, event)
					}
				}
			}
		}(i)
	}

	// Consumers

	pats := append(corpus.MatchingPatterns, corpus.OtherPatterns...)

	for i := 0; i < numConsumers; i++ {
		go func(i int) {
			var (
				c = make(chan string)
			)
			for {
				pat := pats[rand.Intn(len(pats))]
				if err := r.Consume(ctx, []string{pat}, c); err == nil {
					break
				}
				// Bad pattern (which is sort of expected).
				// log.Printf("consumer %d %s", i, pat)
			}

		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				case <-c:
					// Got an event
					// log.Printf("%d got %s", i, event)
					atomic.AddUint64(&forwarded, 1)
				}
			}

			if err := r.StopConsuming(ctx, c); err != nil {
				t.Fatal(err)
			}

		}(i)
	}

	wg.Wait()

	log.Printf("total forwarded %d", atomic.LoadUint64(&forwarded))
}
