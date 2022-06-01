package gen

// A quick router sketch.  Belongs elsewhere.  Far, far from a
// production design.

import (
	"context"
	"reflect"
	"time"

	"github.com/timbray/quamina/core"
	"github.com/timbray/quamina/pruner"
)

// Router receives events and forwards them to the consumers who are
// interested.
//
// When an event arrives, the router forwards that event to every
// consumer with a pattern that matches that event.
type Router struct {
	m *pruner.Matcher

	// patience should be used to identify with slow consumers
	// when they become a problem.
	patience time.Duration
}

func NewRouter() *Router {
	return &Router{
		m:        pruner.NewMatcher(nil),
		patience: time.Second,
	}
}

// Consume should be called by a consumer (for example a handler for a
// new consumer session).
func (r *Router) Consume(ctx context.Context, patterns []string, out chan string) error {
	for i, p := range patterns {
		if err := r.m.AddPattern(out, p); err != nil {
			if 0 < i {
				if e := r.m.DeletePattern(out); e != nil {
					// Sad.  Do something.
				}
			}
			return err
		}
	}
	return nil
}

// StopConsuming should be called by the consumer when it's done.
//
// The router can terminate a consumer independently by closing the
// consumer's channel.
func (r *Router) StopConsuming(ctx context.Context, out chan string) error {
	return r.m.DeletePattern(out)
}

// Route takes an in-bound event and forwards it to the interested
// consumers.
//
// A router will typically consumer a small set of event streams
// (ordered).  Each stream would have its own Flattener.
func (r *Router) Route(ctx context.Context, f core.Flattener, event string) error {
	fs, err := f.Flatten([]byte(event))
	if err != nil {
		return err
	}
	xs, err := r.m.MatchesForFields(fs)
	if err != nil {
		return err
	}

	cs := make(map[reflect.Value]nothing, len(xs))
	for _, x := range xs {
		cs[reflect.ValueOf(x.(chan string))] = na
	}

	// Maybe too fancy and slow.

	var (
		v    = reflect.ValueOf(event)
		todo = len(cs)
		na   = reflect.ValueOf(nil)
		done = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}
	)

	for 0 < len(cs) {
		cases := make([]reflect.SelectCase, 0, todo+2)
		cases = append(cases, done)
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(time.NewTimer(r.patience).C),
		})
		for c := range cs {
			if c == na {
				continue
			}
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: c,
				Send: v,
			})
		}
		i, _, _ := reflect.Select(cases)
		switch i {
		case 0:
			return context.Canceled
		case 1:
			// All of the damn consumers are slow.
		}
		if _, have := cs[cases[i].Chan]; !have {
			panic(cases[i].Chan)
		}

		delete(cs, cases[i].Chan)
		todo--
	}

	return nil
}
