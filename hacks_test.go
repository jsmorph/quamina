package quamina

import (
	"encoding/json"
	"testing"
)

func TestHacksUse(t *testing.T) {
	parse := func(js string) any {
		var x any
		if err := json.Unmarshal([]byte(js), &x); err != nil {
			t.Fatal(err)
		}
		return x
	}
	y, err := UseHacks(parse(`{"likes":[{"numeric":["<",42]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	pat, err := niceJSON(y)
	if err != nil {
		t.Fatal(err)
	}

	q, _ := New()
	if err := q.AddPattern(1, pat); err != nil {
		t.Fatal(err)
	}
	xs, err := q.MatchesForEvent([]byte(`{"likes":22}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(xs) == 0 {
		t.Fatal(xs)
	}

}

func TestHacksMatch(t *testing.T) {
	f := func(t *testing.T, pat, event string, wantMatch bool) {
		q, _ := New()
		if err := q.AddPattern(1, pat); err != nil {
			t.Fatal(err)
		}
		xs, err := q.MatchesForEvent([]byte(event))
		if err != nil {
			t.Fatal(err)
		}
		if wantMatch {
			if 0 == len(xs) {
				t.Fatal(xs)
			}
		} else {
			if 0 < len(xs) {
				t.Fatal(xs)
			}
		}
	}

	t.Run("eval", func(t *testing.T) {
		t.Skip()
		f(t,
			`{"likes":["!function(v) { return v.length > 3 }"]}`,
			`{"likes":"tacos"}`, true)
	})

	t.Run("numeric-match-simple", func(t *testing.T) {
		f(t,
			`{"likes":["{\"numeric\":[\"<\",42]}"]}`,
			`{"likes":12}`,
			true)
	})

	t.Run("numeric-nomatch-two", func(t *testing.T) {
		f(t,
			`{"likes":["{\"numeric\":[\"<\",42,\">\",52]}"]}`,
			`{"likes":62}`,
			false)
	})

	t.Run("numeric-match-near", func(t *testing.T) {
		f(t,
			`{"likes":["{\"numeric\":[\"~=\",42.000000000]}"]}`,
			`{"likes":                       42.000000001}`,
			true)
	})

	t.Run("numeric-nomatch-near", func(t *testing.T) {
		f(t,
			`{"likes":["{\"numeric\":[\"~=\",42.0000000]}"]}`,
			`{"likes":                       42.0000001}`,
			false)
	})
}
