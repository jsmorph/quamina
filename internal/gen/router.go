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

	// Maybe too fancy and slow.

	v := reflect.ValueOf(event)

	cases := make([]reflect.SelectCase, 0, len(xs)+2)

	// We can get canceled.
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	})

	// We won't wait all day.
	cases = append(cases, reflect.SelectCase{
		Dir: reflect.SelectRecv,
		// Chan: // To fill in below.
	})

	for _, x := range xs {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(x.(chan string)),
			Send: v,
		})
	}

LOOP:
	for todo := len(xs); 0 < todo; {
		// Get a new timer.
		cases[1].Chan = reflect.ValueOf(time.NewTimer(r.patience).C)

		i, _, _ := reflect.Select(cases)
		switch i {
		case 0:
			return context.Canceled
		case 1:
			// Timer fired.
			//
			// All of the (remaining) damn consumers are
			// slow.  Do something?  We should probably
			// terminate those consumers (by closing their
			// channels).  Make sure our timer is
			// resonable.
			for _, c := range cases[2:] {
				close(c.Chan.Interface().(chan string))
			}
			break LOOP
		default:
			// Remove that case and then carry on.
			cases[i] = cases[len(cases)-1]
			cases = cases[0 : len(cases)-1]
			todo--
		}
	}

	return nil
}
