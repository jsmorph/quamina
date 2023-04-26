package quamina

import (
	"encoding/json"
	"fmt"
	"strings"
)

func AddExtendedPattern(q *Quamina, id any, pat string) error {
	predicateFields := []string{"numeric", "equals-insensitive"}
	return q.AddPattern(id, UsingExtension(pat, predicateFields...))
}

func UseStdExtension() error {

	ThePredicateParser = func(spec []byte) (Predicate, error) {

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

	return nil
}
