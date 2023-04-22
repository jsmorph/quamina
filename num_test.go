package quamina

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func TestNumber(t *testing.T) {
	x := []byte("123")
	// x = append(x, valueTerminator)

	A0 := &dfaStep{table: newSmallTable[*dfaStep]()}
	A1 := &dfaStep{table: newSmallTable[*dfaStep]()}
	A2 := &dfaStep{table: newSmallTable[*dfaStep]()}
	A3 := &dfaStep{table: newSmallTable[*dfaStep]()}
	A0.table.addByteStep(x[0], A1)
	A1.table.addByteStep(x[1], A2)
	A2.table.addByteStep(x[2], A3)

	AFM := newFieldMatcher()
	AFM.fields().transitions[""] = newValueMatcher()
	st := newDfaTransition(AFM)
	A3.table.addByteStep(valueTerminator, st)

	state := &vmFields{startDfa: A0.table}
	vm := newValueMatcher()
	vm.update(state)
	matches := vm.transitionOn(x)

	log.Printf("%#v", matches)
}

func New0NumericAutomaton(x []byte) *dfaStep {
	zero := int('0')
	s0 := &dfaStep{table: newSmallTable[*dfaStep]()}
	s1 := &dfaStep{table: newSmallTable[*dfaStep]()}
	s0.table.addRangeSteps(zero, int(x[0]), s1)

	AFM := newFieldMatcher()
	AFM.fields().transitions[""] = newValueMatcher()
	st := newDfaTransition(AFM)
	s1.table.addByteStep(valueTerminator, st)

	return s0
}

func New1NumericAutomaton(x []byte) *dfaStep {
	zero := int('0')
	s0 := &dfaStep{table: newSmallTable[*dfaStep]()}
	s := s0
	for _, b := range x {
		lt := &dfaStep{table: newSmallTable[*dfaStep]()}
		s.table.addRangeSteps(zero, int(b), lt)
		s = lt
	}

	AFM := newFieldMatcher()
	AFM.fields().transitions[""] = newValueMatcher()
	st := newDfaTransition(AFM)
	s.table.addByteStep(valueTerminator, st)

	return s0
}

var (
	zero = int('0')
	ten  = int('9') + 1
)

func notLongerThan(x []byte) *dfaStep {
	s := &dfaStep{table: newSmallTable[*dfaStep]()}
	s.table.addByteStep(valueTerminator, moveOn())
	if 0 < len(x) {
		s.table.addRangeSteps(zero, ten, notLongerThan(x[1:]))
	}

	return s
}

func shorterThan(x []byte) *dfaStep {
	s := &dfaStep{table: newSmallTable[*dfaStep]()}
	if 0 < len(x) {
		s.table.addByteStep(valueTerminator, moveOn())
		s.table.addRangeSteps(zero, ten, shorterThan(x[1:]))
	}

	return s
}

func NewNumericAutomaton(x []byte) *dfaStep {
	s := &dfaStep{table: newSmallTable[*dfaStep]()}
	if 0 < len(x) {
		s.table.addByteStep(valueTerminator, moveOn())
	}

	if len(x) == 0 {
		return s
	}

	var (
		d    = int(x[0])
		rest = x[1:]
	)

	s.table.addRangeSteps(zero, d, notLongerThan(rest))
	s.table.addByteStep(byte(d), NewNumericAutomaton(rest))
	s.table.addRangeSteps(d+1, ten, shorterThan(rest))

	return s
}

func moveOn() *dfaStep {
	s := &dfaStep{table: newSmallTable[*dfaStep]()}
	afm := newFieldMatcher()
	afm.fields().transitions[""] = newValueMatcher()
	st := newDfaTransition(afm)
	s.table.addByteStep(valueTerminator, st)
	return s
}

func testOne(t *testing.T, x, y int) {
	var (
		xb    = []byte(strconv.Itoa(x))
		yb    = []byte(strconv.Itoa(y))
		match = y < x
	)
	yb = append(yb, valueTerminator)

	s := NewNumericAutomaton(xb)

	state := &vmFields{startDfa: s.table}
	vm := newValueMatcher()
	vm.update(state)
	matches := vm.transitionOn(yb)
	if match {
		if len(matches) == 0 {
			t.Fatal(x, y, match, len(matches))
		}
	} else {
		if 0 < len(matches) {
			t.Fatal(x, y, match, len(matches))
		}
	}
}

func TestThis(t *testing.T) {
	testOne(t, 30, 40)
}

func TestThat(t *testing.T) {
	type Case struct {
		x, y  string
		match bool
	}

	for i := 0; i < 1000; i++ {
		var (
			x = rand.Intn(100)
			y = rand.Intn(100)
		)
		t.Run(fmt.Sprintf("%d < %d: %v", y, x, y < x), func(t *testing.T) {
			testOne(t, x, y)
		})
	}
}
