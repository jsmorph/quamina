package gen

import (
	"fmt"
	quamina "quamina/lib"
)

// PatternIndex has implementations based on this Matcher as well as a
// slightly augmented (here) core quamina.Matcher.
//
// Can be useful in testing.
type PatternIndex interface {
	AddPattern(quamina.X, string) error
	DelPattern(quamina.X) (bool, error)
	MatchesForJSONEvent(event []byte) ([]quamina.X, error)
	Rebuild(bool) error
}

// CoreMatcher wraps a quamina.Matcher with addtional methods to
// support the PatternIndex interface.  Those additional methods
// always return the error NotImplemented.
type CoreMatcher struct {
	*quamina.Matcher
}

func NewCoreMatcher() *CoreMatcher {
	return &CoreMatcher{
		Matcher: quamina.NewMatcher(),
	}
}

var NotImplemented = fmt.Errorf("not implemented")

func (m *CoreMatcher) DelPattern(quamina.X) (bool, error) {
	return false, NotImplemented
}

func (m *CoreMatcher) Rebuild(fearlessly bool) error {
	return NotImplemented
}
