package quamina

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

type Predicate func([]byte) bool

type extensionMatcher struct {
	Predicate  Predicate
	Transition *fieldMatcher
}

type PredicateParser func([]byte) (Predicate, error)

var ThePredicateParser PredicateParser

func makeExtensionMatcher(spec []byte) (*extensionMatcher, error) {
	log.Printf("debug makeExtensionMatcher <%s>", spec)
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

func toJSON(x interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(x)
	return buffer.Bytes(), err
}

func readExtensionSpecial(pb *patternBuild, valsIn []typedVal) (pathVals []typedVal, err error) {
	log.Printf("debug readExtensionSpecial %v", valsIn)

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
