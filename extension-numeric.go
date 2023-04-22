package quamina

// This code is supposed to implement roughly EventBridge's numeric
// suite.  Roughly.
//
// See extension_test.go for examples.

import (
	"fmt"
	"math"
)

var (
	// FloatNearMaxDelta is used by ~= (NumRelNear) to provide
	// floating point approximate equality.
	FloatNearMaxDelta = 1e-08
)

type NumericRel int64

const (
	NumRelLT NumericRel = iota
	NumRelGT
	NumRelLTE
	NumRelGTE
	NumRelEq
	NumRelNE
	NumRelNear
)

type NumericConstraint struct {
	Rel NumericRel
	Arg float64
}

type NumericConstraints []NumericConstraint

func CompileNumericConstraint(rel any, arg any) (*NumericConstraint, error) {
	s, is := rel.(string)
	if !is {
		return nil, fmt.Errorf("bad rel %v (%T)", rel, rel)
	}

	var r NumericRel
	switch s {
	case "<":
		r = NumRelLT
	case ">":
		r = NumRelGT
	case "<=":
		r = NumRelLTE
	case ">=":
		r = NumRelGTE
	case "==", "=":
		r = NumRelEq
	case "!=", "<>":
		r = NumRelNE
	case "~=":
		r = NumRelNear
	default:
		return nil, fmt.Errorf("unknown numeric relation %v", rel)
	}

	var y float64
	switch vv := arg.(type) {
	case float64:
		y = vv
	case int64:
		y = float64(vv)
	case int:
		y = float64(vv)
	default:
		return nil, fmt.Errorf("%v (%T) isn't numeric enough", arg, arg)
	}

	return &NumericConstraint{r, y}, nil
}

func CompileNumericConstraints(spec []any) (NumericConstraints, error) {
	n := len(spec)
	if n == 0 {
		return nil, fmt.Errorf("need at least one numeric constraint")
	}
	if n%2 != 0 {
		return nil, fmt.Errorf("odd number of args: %d", n)
	}

	cs := make(NumericConstraints, 0, n/2)

	for i := 0; i < n; i += 2 {
		c, err := CompileNumericConstraint(spec[i], spec[i+1])
		if err != nil {
			return nil, err
		}
		cs = append(cs, *c)
	}

	return cs, nil
}

func (c NumericConstraints) Matches(x float64) (bool, error) {
	for _, inst := range c {
		var ok bool
		switch inst.Rel {
		case NumRelLT:
			ok = x < inst.Arg
		case NumRelGT:
			ok = x > inst.Arg
		case NumRelLTE:
			ok = x <= inst.Arg
		case NumRelGTE:
			ok = x >= inst.Arg
		case NumRelEq:
			ok = x == inst.Arg
		case NumRelNE:
			ok = x != inst.Arg
		case NumRelNear:
			ok = math.Abs(x-inst.Arg) < FloatNearMaxDelta
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
