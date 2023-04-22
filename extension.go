package quamina

// Danger zone: This file provides support for external matching
// predicates.
//
// Quamina is fast.  This gear is likely not.  Use at your own risk.
// Perhaps a good use is custom experimentation without having to
// figure out automata.
//
// You have been warned.

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ThePredicateParser, if not nil, provides a parser for a "extension"
// constraint.
//
// Example pattern: {"likes":[{"extension":{"string-of-length":42}}]}
//
// In that example, {"string-of-length":42"} would be given to this
// PredicateParser to be parsed into a Predicate.
//
// This variable is a global variable because I can't think of a
// better place to put it at the moment.  Aside: It's easy to imagine
// a collection of configuration options which together should be
// carried around when running Quamina functions/methods.  Something
// like
//
// type Cfg struct { PredicateParser, FooMode, BarSwitch }
//
// func (cfg *Cfg) NewQ() (Q,error) ...
var ThePredicateParser PredicateParser

// Predicate is the core externally-provided type to determine whether
// given bytes is acceptable for matching.
type Predicate func([]byte) bool

type PredicateParser func([]byte) (Predicate, error)

// extensionMatcher is functionally similar to an automaton that can
// match a value.
type extensionMatcher struct {
	Predicate  Predicate
	Transition *fieldMatcher
}

// makeExtensionMatcher uses ThePredicateParser to construct an
// extensionMatcher and its Predicate.
func makeExtensionMatcher(spec []byte) (*extensionMatcher, error) {
	if ThePredicateParser == nil {
		return nil, fmt.Errorf("ThePredicateParser is nil")
	}
	p, err := ThePredicateParser(spec)
	if err != nil {
		return nil, fmt.Errorf("ThePredicateParser returned an error when attempting to parse %s: %w", spec, err)
	}

	return &extensionMatcher{
		Predicate:  p,
		Transition: newFieldMatcher(),
	}, nil
}

// toJSON exists to serialize JSON without escaping < and >.
func toJSON(x interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(x)
	return bytes.TrimSpace(buffer.Bytes()), err
}

// readExtensionSpecial is like readExistsSpecial and other read*Specials.
func readExtensionSpecial(pb *patternBuild, valsIn []typedVal) (pathVals []typedVal, err error) {
	var x any
	if err = pb.jd.Decode(&x); err != nil {
		err = fmt.Errorf("value for 'extension' must be ... something: %w", err)
		return
	}

	bs, err := toJSON(&x)
	if err != nil {
		err = fmt.Errorf("internal error dealing with 'extension': %w", err)
		return
	}

	val := typedVal{
		vType: extensionType,
		val:   string(bs),
	}
	pathVals = append(pathVals, val)

	// has to be } or tokenizer will throw error
	_, err = pb.jd.Token()

	return
}

// UsingExtension is a shameful utility that demotes given properties
// to "extension" values.
//
// Example:
//
//	{"likes":[{"numeric":["<",42]}]}
//
// becomes
//
//	{"likes":{"extension":{"numeric":["<",42]}}}
//
// when given "numeric" as one of the props.
//
// This utility allows a caller to pretend that extensions are
// supported inline.
func UsingExtension(pat string, props ...string) string {
	var x any
	if err := json.Unmarshal([]byte(pat), &x); err != nil {
		return pat
	}
	y := usingExtensionParsed(x, props...)
	bs, err := toJSON(&y)
	if err != nil {
		return pat
	}

	return string(bs)
}

func usingExtensionParsed(x any, props ...string) any {
	switch vv := x.(type) {
	case []any:
		acc := make([]any, 0, len(vv))
		if len(vv) == 1 {
			switch vv := vv[0].(type) {
			case map[string]any:
				if len(vv) == 1 {
					for _, prop := range props {
						if spec, has := vv[prop]; has {
							x := map[string]any{
								"extension": map[string]any{
									prop: spec,
								},
							}
							return append(acc, x)
						}
					}
				}
			}
		}

		for _, y := range vv {
			acc = append(acc, y)
		}

		return acc
	case map[string]any:
		acc := make(map[string]any, len(vv))
		for p, v := range vv {
			acc[p] = usingExtensionParsed(v, props...)
		}
		return acc
	default:
		return x
	}
}
