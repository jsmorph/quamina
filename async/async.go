package async

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/timbray/quamina/core"
)

// ToDo: Enforce Matcher.limit?  What about other limits?

type (
	stringSet map[string]nothing

	patternSet map[core.X]stringSet

	nothing struct{}
)

var na = nothing{}

func (ss stringSet) Add(s string) {
	ss[s] = na
}

func (s patternSet) Add(x core.X, p string) {
	ps, have := s[x]
	if !have {
		ps = make(stringSet)
		s[x] = ps
	}
	ps.Add(p)
}

func (s patternSet) Delete(x core.X) {
	delete(s, x)
}

func (s patternSet) Size() int {
	count := 0
	for _, ps := range s {
		count += len(ps)
	}
	return count
}

type mutation struct {
	X        core.X
	Pattern  string
	Deletion bool
}

func (s patternSet) Mutate(m *mutation) {
	if m.Deletion {
		s.Delete(m.X)
	} else {
		s.Add(m.X, m.Pattern)
	}
}

// Matcher is (almost) a Quamina Matcher that operates by mutating the
// core state asychronously.
//
// The "almost" qualification is due to a context.Context parameter
// that each primary method takes as a first argument.  This context
// is required due to channel communications with the background
// process that performs the periodic rebuilds.
//
// By default, the Matcher rebuilds the core Matcher every one second.
// This work is performed in another goroutine, so Matcher interface
// method callers should not see (much) additional latency due to
// rebuilds.  The cost is that mutating operations (AddPattern and
// DeletePattern) do not become effective immediately.  The delay is
// dependent on a configurable rebuild initiation policy as well as
// the size and complexity of the live set of patterns.
//
// An application using this Matcher will see CPU consumed
// asynchronous for from-scratch rebuilds.  In addition, the
// application will see additional heap usage as the new core Matcher
// is rebuilt while the previous core Matcher remains in place during
// that work.
type Matcher struct {
	RebuildReports chan RebuildReport
	Logging        bool
	mutations      chan mutation
	core           atomic.Value
	limit          int

	live patternSet
}

type RebuildReport struct {
	T           time.Time
	Duration    time.Duration
	LiveSizeSet int
	Err         error
}

var defaultMutationsInitialCapacity = 1024

// NewMatcher makes a new Matcher!
func NewMatcher() *Matcher {
	m := &Matcher{
		mutations: make(chan mutation),
		limit:     defaultMutationsInitialCapacity,
		live:      make(patternSet),
	}

	m.core.Store(core.NewCoreMatcher())

	return m
}

func (m *Matcher) logf(format string, args ...interface{}) {
	if !m.Logging {
		return
	}
	log.Printf(format, args...)
}

func (m *Matcher) matcher() core.Matcher {
	return m.core.Load().(core.Matcher)
}

// Policy implements a rebuild trigger policy, which can be based (if
// desired) on the number or type of mutations accumulated since the
// previously triggered rebuild.
//
// The returned channel C provides time.Times only to make it easy to
// use a time.Ticker as a Policy.
type Policy interface {
	C() <-chan time.Time
	Mutation(*mutation)
}

// Run executes asynchronous, more or less continuous rebuilds based
// on the given Policy.
//
// If the given Policy is nil, DefaultPolicy will be used.
//
// This method executes in the current goroutine, so callers should
// almost always execute this method in a new goroutine.
func (m *Matcher) Run(ctx context.Context, p Policy) error {
	if p == nil {
		p = DefaultPolicy
	}

	m.logf("running")
	var (
		trigger    = p.C()
		rebuilding bool
		errs       = make(chan error)
		muts       = make([]mutation, 0, m.limit)
	)
	for {
		select {
		case <-ctx.Done():
			return nil
		case mut := <-m.mutations:
			m.logf("heard mutation %#v", mut)
			p.Mutation(&mut)
			muts = append(muts, mut)
		case <-trigger:
			m.logf("trigger")
			if rebuilding {
				// Do not try to start a rebuild when
				// one is in progress.  Yes, queuing
				// problems could arise, and the
				// primary consequence would be
				// increasing lag for mutations to
				// take effect.
				continue
			}
			rebuilding = true

			// Make a copy of the mutations.  Since this
			// work is performed in this thread, it will
			// block AddPattern and DeletePattern
			// operations since the mutations channel is
			// currently unbuffered.  Maybe (or maybe not)
			// buffer that channel.

			acc := make([]mutation, len(muts))
			copy(acc, muts)
			muts = make([]mutation, 0, m.limit)

			// Start the rebuild in a new goroutine (for
			// now).  This loop will hear when the rebuild
			// is complete via the errs channel.
			go m.rebuild(ctx, acc, errs)
		case err := <-errs:
			// A rebuild terminated.
			m.logf("rebuilt (%v)", err)
			rebuilding = false
			if err != nil {
				return err
			}
		}
	}
}

// DefaultPolicy is what a Matcher.Run will use if no Policy is provided.
var DefaultPolicy = &TempoPolicy{time.Second}

// rebuild performns the actual core Matcher rebuild.
//
// This method is currently performed in a new goroutine by
// Matcher.Run.
func (m *Matcher) rebuild(ctx context.Context, muts []mutation, errs chan error) {

	if 0 == len(muts) {
		// No mutations, so nothing to do.
		return
	}

	t := time.Now().UTC()

	// Update the previous set of live patterns.
	for _, mut := range muts {
		m.live.Mutate(&mut)
	}

	var (
		c   = core.NewCoreMatcher()
		err error
	)

	// Build a new core Matcher from scratch based on the live
	// patterns.
ADDS:
	for x, ps := range m.live {
		for p := range ps {
			if e := c.AddPattern(x, p); e != nil {
				err = e
				break ADDS
			}
		}
	}

	if err == nil {
		// Atomically update the in-use core Matcher.
		m.core.Store(c)
	}

	select {
	case <-ctx.Done():
	case errs <- err:
		// err can be nil to signal we are done.
	}

	if m.RebuildReports != nil {
		select {
		case <-ctx.Done():
		case m.RebuildReports <- RebuildReport{
			T:           t,
			Duration:    time.Now().Sub(t),
			Err:         err,
			LiveSizeSet: m.live.Size(),
		}:
		}
	}
}

func (m *Matcher) AddPattern(ctx context.Context, x core.X, pat string) error {

	// Determine now if the AddPattern will likely work (so we can
	// some asynchronous AddPattern user errors).  This operation
	// sure seems wasteful, but we'd like to report user errors to
	// the caller now rather than deal with them later during
	// asynchronous processing that dissociated from the caller of
	// this method.  ToDo: Something better?

	if true {
		// Hopefully we get a lot of stack allocation ...
		check := core.NewCoreMatcher()
		if err := check.AddPattern(x, pat); err != nil {
			return err
		}
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case m.mutations <- mutation{
		Pattern: pat,
		X:       x,
	}:
		return nil
	}

}

func (m *Matcher) MatchesForJSONEvent(ctx context.Context, event []byte) ([]core.X, error) {
	return m.matcher().MatchesForJSONEvent(event)
}

func (m *Matcher) MatchesForFields(ctx context.Context, fields []core.Field) ([]core.X, error) {
	return m.matcher().MatchesForFields(fields)
}

func (m *Matcher) DeletePattern(ctx context.Context, x core.X) error {
	select {
	case <-ctx.Done():
		return context.Canceled
	case m.mutations <- mutation{
		X:        x,
		Deletion: true,
	}:
		return nil
	}
}
