package gen

import (
	"testing"

	"quamina/pruner"
)

func TestIndex(t *testing.T) {
	var i PatternIndex

	i = NewCoreMatcher()

	i = pruner.NewMatcher(nil)

	if err := i.AddPattern(1, `{"likes":["queso"]}`); err != nil {
		t.Fatal(err)
	}
}
