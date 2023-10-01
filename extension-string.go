package quamina

// This code is supposed to implement roughly EventBridge's string
// suite.  Roughly.
//
// See extension_test.go for examples.

import (
	"fmt"
)

type StringRel int64

const (
	StrRelLT StringRel = iota
	StrRelGT
	StrRelLTE
	StrRelGTE
	StrRelEq
	StrRelNE
	StrRelNear
)

type StringConstraint struct {
	Rel StringRel
	Arg string
}

type StringConstraints []StringConstraint

func CompileStringConstraint(rel any, arg any) (*StringConstraint, error) {
	s, is := rel.(string)
	if !is {
		return nil, fmt.Errorf("bad rel %v (%T)", rel, rel)
	}

	var r StringRel
	switch s {
	case "<":
		r = StrRelLT
	case ">":
		r = StrRelGT
	case "<=":
		r = StrRelLTE
	case ">=":
		r = StrRelGTE
	case "==", "=":
		r = StrRelEq
	case "!=", "<>":
		r = StrRelNE
	case "~=":
		r = StrRelNear
	default:
		return nil, fmt.Errorf("unknown string relation %v", rel)
	}

	var y string
	switch vv := arg.(type) {
	case string:
		y = vv
	default:
		return nil, fmt.Errorf("%v (%T) isn't a string", arg, arg)
	}

	return &StringConstraint{r, y}, nil
}

func CompileStringConstraints(spec []any) (StringConstraints, error) {
	n := len(spec)
	if n == 0 {
		return nil, fmt.Errorf("need at least one string constraint")
	}
	if n%2 != 0 {
		return nil, fmt.Errorf("odd number of args: %d", n)
	}

	cs := make(StringConstraints, 0, n/2)

	for i := 0; i < n; i += 2 {
		c, err := CompileStringConstraint(spec[i], spec[i+1])
		if err != nil {
			return nil, err
		}
		cs = append(cs, *c)
	}

	return cs, nil
}

func (c StringConstraints) Matches(x string) (bool, error) {
	for _, inst := range c {
		var ok bool
		switch inst.Rel {
		case StrRelLT:
			ok = x < inst.Arg
		case StrRelGT:
			ok = x > inst.Arg
		case StrRelLTE:
			ok = x <= inst.Arg
		case StrRelGTE:
			ok = x >= inst.Arg
		case StrRelEq:
			ok = x == inst.Arg
		case StrRelNE:
			ok = x != inst.Arg
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
