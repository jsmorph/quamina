package quamina

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// predicateFields are the custom fields that aPredicateParser supports.
var predicateFields = []string{"numeric", "weird", "equals-insensitive"}

// aPredicateParser is a demo/test PredicateParser that supports some
// matching extensions.
var aPredicateParser = func(spec []byte) (Predicate, error) {

	// Generic supports some demo matching extension.  This struct
	// provides a bit of type safety and facilitates parsing
	// (assuming some absurd JSON gymastics).
	type Generic struct {
		// Numeric supposedly roughly supports the full
		// EventBridge "numeric" suite.
		Numeric []any `json:"numeric,omitempty"`

		// Weird is a predicate the returns true when given a
		// string with length of the given value.
		//
		// The type is a pointer so we can detect when it's
		// actually provided (since zero is a valid length).
		Weird *int `json:"weird,omitempty"`

		// EqualsInsensitive provides case-insenstive string
		// matching.  Note that the value is an array, which
		// is different from EventBridge.  Also note that the
		// property name is different, too.  I did not
		// overthink this stuff.
		EqualsInsensitive []string `json:"equals-insensitive,omitempty"`
	}

	var g Generic
	// Yeah, this code runs when a pattern is compiled.
	if err := json.Unmarshal(spec, &g); err != nil {
		return nil, fmt.Errorf("PredicateParser failed to parse %s: %w", spec, err)
	}

	{
		// Insist that we have only one field.
		count := 0
		if g.Numeric != nil {
			count++
		}
		if g.Weird != nil {
			count++
		}
		if g.EqualsInsensitive != nil {
			count++
		}

		if count != 1 {
			return nil, fmt.Errorf("PredicateParser requires exactly one field")
		}
	}

	if g.Numeric != nil {
		nc, err := CompileNumericConstraints(g.Numeric)
		if err != nil {
			return nil, fmt.Errorf("PredicateParser failed to compile numeric %s: %w", spec, err)
		}

		return func(bs []byte) bool {
			var x float64
			if err := json.Unmarshal(bs, &x); err != nil {
				return false
			}

			matches, err := nc.Matches(x)
			if err != nil {
				return false
			}
			return matches
		}, nil
	}

	if g.Weird != nil {
		return func(bs []byte) bool {
			var s string
			if err := json.Unmarshal(bs, &s); err != nil {
				return false
			}
			return len(s) == *g.Weird
		}, nil
	}

	if g.EqualsInsensitive != nil {
		for i, s := range g.EqualsInsensitive {
			g.EqualsInsensitive[i] = strings.ToLower(s)
		}
		return func(bs []byte) bool {
			var s string
			if err := json.Unmarshal(bs, &s); err != nil {
				return false
			}
			s = strings.ToLower(s)
			for _, allowed := range g.EqualsInsensitive {
				if allowed == s {
					return true
				}
			}
			return false
		}, nil
	}

	return nil, fmt.Errorf("PredicateParser requires exactly one field")
}

func TestExtensionMatching(t *testing.T) {

	ThePredicateParser = aPredicateParser
	q, err := New()
	if err != nil {
		t.Fatal(err)
	}

	add := func(id int, pat string) {
		pat = UsingExtension(pat, predicateFields...)
		err = q.AddPattern(id, pat)
		if err != nil {
			t.Fatalf("AddPattern error on %s: %s", pat, err)
		}
	}

	check := func(s string, want ...int) {
		t.Run("", func(t *testing.T) {
			xs, err := q.MatchesForEvent([]byte(s))
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[int]bool)
			for _, x := range xs {
				got[x.(int)] = true
			}
			for _, wanted := range want {
				if _, have := got[wanted]; !have {
					t.Fatalf("wanted %d", wanted)
				}
				delete(got, wanted)
			}
			if 0 < len(got) {
				t.Fatalf("didn't want %v with %s", got, s)
			}
		})
	}

	add(1, `{"likes":[{"numeric":["<",10]}]}`)

	check(`{"likes":5}`, 1)
	check(`{"likes":15}`)

	add(1, `{"likes":[{"numeric":[">",20]}]}`)

	check(`{"likes":5}`, 1)
	check(`{"likes":15}`)
	check(`{"likes":25}`, 1)

	add(2, `{"likes":[{"numeric":[">",20]}]}`)

	check(`{"likes":5}`, 1)
	check(`{"likes":15}`)
	check(`{"likes":25}`, 1, 2)

	add(3, `{"likes":[{"numeric":[">",30]}]}`)

	check(`{"likes":5}`, 1)
	check(`{"likes":15}`)
	check(`{"likes":25}`, 1, 2)
	check(`{"likes":35}`, 1, 2, 3)

	add(4, `{"likes":[{"weird":5}]}`)

	check(`{"likes":5}`, 1)
	check(`{"likes":15}`)
	check(`{"likes":25}`, 1, 2)
	check(`{"likes":35}`, 1, 2, 3)
	check(`{"likes":"tacos"}`, 4)

	add(4, `{"likes":["queso"]}`)

	check(`{"likes":"queso"}`, 4)

	add(5, `{"likes":["queso"]}`)
	check(`{"likes":"queso"}`, 4, 5)
	check(`{"likes":"QUESO"}`, 4)

	add(6, `{"likes":[{"equals-insensitive":["queso"]}]}`)
	check(`{"likes":"QUESO"}`, 4, 6)

	check(`{"needs":"margarita"}`)

	add(6, `{"needs":["margarita"]}`)
	check(`{"needs":"margarita"}`, 6)

	add(7, `{"needs":{"some":["chips"]}}`)
	check(`{"needs":"margarita"}`, 6)

	add(8, `{"needs":[{"numeric":[">",30]}]}`)
	check(`{"needs":"margarita"}`, 6)
	check(`{"needs":40}`, 8)

	if q, err = New(); err != nil {
		t.Fatal(err)
	}

	add(1, `{"needs":[{"numeric":["<",200]}]}`)
	check(`{"needs":1.23e2}`, 1)
	add(2, `{"needs":[{"numeric":[">",100]}]}`)
	check(`{"needs":124}`, 1, 2)
	check(`{"needs":1.24e2}`, 1, 2)

	if q, err = New(); err != nil {
		t.Fatal(err)
	}

	add(1, `{"needs":[{"numeric":["<",200,">",100]}]}`)
	check(`{"needs":1.23e2}`, 1)
	check(`{"needs":1.24e3}`)

}

func TestExtensionImplicit(t *testing.T) {
	type Case struct {
		Pat, Want string
	}
	fields := append(predicateFields, "foo")
	for _, c := range []Case{
		{`{"likes":[{"numeric":["<",10,">",3]}]}`,
			`{"likes":[{"extension":{"numeric":["<",10,">",3]}}]}`,
		},
		{`{"likes":[{"foo":{"likes":"tacos"}}]}`,
			`{"likes":[{"extension":{"foo":{"likes":"tacos"}}}]}`,
		},
		{`{"likes":[{"equals-insensitive":["tacos"]}]}`,
			`{"likes":[{"extension":{"equals-insensitive":["tacos"]}}]}`,
		},
	} {
		t.Run("", func(t *testing.T) {
			got := UsingExtension(c.Pat, fields...)
			if c.Want != got {
				t.Fatalf("want %s != %s got", c.Want, got)
			}
		})
	}
}

func BenchmarkNumericExtension(b *testing.B) {
	ThePredicateParser = aPredicateParser

	q, err := New()
	if err != nil {
		b.Fatal(err)
	}

	pat := `{"likes":[{"numeric":["<",10]}]}`
	msg := `{"likes":5}`

	if err := q.AddPattern(1, UsingExtension(pat, "numeric")); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := q.MatchesForEvent([]byte(msg)); err != nil {
			b.Fatal(err)
		}
	}
}
