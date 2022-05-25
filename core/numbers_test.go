package core

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
)

func TestVariants(t *testing.T) {
	f := []string{
		"350",
		"350.0",
		"350.0000000000",
		"3.5e2",
	}
	var o []string
	for _, s := range f {
		c, err := canonicalize([]byte(s))
		if err != nil {
			t.Errorf("canon err on %s: %s", s, err.Error())
		}
		o = append(o, c)
	}
	for i := 1; i < len(o); i++ {
		if o[i] != o[i-1] {
			t.Errorf("%s and %s differ", o[i-1], o[i])
		}
	}
}

func TestOrdering(t *testing.T) {
	var in []float64
	for i := 0; i < 10000; i++ {
		f := rand.Float64() * math.Pow(10, 9) * 2
		f -= nineDigits
		in = append(in, f)
	}
	sort.Float64s(in)
	var out []string
	for _, f := range in {
		s := fmt.Sprintf("%f", f)
		c, err := canonicalize([]byte(s))
		if err != nil {
			t.Errorf("failed on %s", s)
		}
		out = append(out, c)
	}
	if !sort.StringsAreSorted(out) {
		t.Errorf("Not sorted")
	}
	for i, c := range out {
		if len(c) != 19 {
			t.Errorf("%s: %d at %d", c, len(c), i)
		}
	}
}
