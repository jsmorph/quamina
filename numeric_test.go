package quamina

import (
	"fmt"
	"log"
	"strconv"
	"testing"
)

type State struct {
	// -1 means "any"
	Note string
	Ts   map[int]*State
}

func NotLonger(n int) *State {
	if n == 0 {
		return &State{
			Note: "match",
		}
	}
	ts := make(map[int]*State)
	ts[-1] = NotLonger(n - 1)
	return &State{
		Note: fmt.Sprintf("NotLonger(%d)", n),
		Ts:   ts,
	}
}

func Longer(n int) *State {
	if n <= 0 {
		return nil // Failure.
	}
	ts := make(map[int]*State)
	ts[-1] = Longer(n - 1)
	return &State{
		Note: fmt.Sprintf("Longer(%d)", n),
		Ts:   ts,
	}
}

func GenState(n int) *State {
	var (
		ts      = make(map[int]*State)
		y       = strconv.Itoa(n)
		ylen    = len(y)
		y0, err = strconv.Atoi(y[0:1])
	)

	if err != nil {
		panic(err)
	}

	for x0 := 0; x0 < 10; x0++ {
		log.Printf("GenState n: %v, x0: %v, y0: %v", n, x0, y0)

		if x0 < y0 {
			ts[x0] = NotLonger(ylen - 1)
		} else if x0 > y0 {
			ts[x0] = Longer(ylen - 1)
		} else { // x0 == y0
			if ylen == 1 {
				ts[x0] = nil
			} else {
				more, err := strconv.Atoi(y[1:])
				if err != nil {
					panic(err)
				}
				ts[x0] = GenState(more)
			}
		}
	}

	return &State{
		Ts: ts,
	}
}

func (s *State) Step(n string) *State {
	log.Printf("Step x: %v", n)
	x0, err := strconv.Atoi(n[0:1])
	if err != nil {
		panic(err)
	}

	if next, have := s.Ts[x0]; have {
		return next
	}

	if next, have := s.Ts[-1]; have {
		return next
	}

	return nil
}

func (s *State) Run(n string) *State {
	for {
		s = s.Step(n)
		if s == nil {
			return nil
		}
		if n = n[1:]; len(n) == 0 {
			return s
		}
	}
}

func TestNumericLT(t *testing.T) {

	type Case struct {
		y     string
		match bool
	}

	s := GenState(42)
	for _, c := range []Case{{"31", true}, {"3", true}, {"40", true}, {"41", true}, {"42", false}, {"52", false}, {"111", false}} {
		log.Printf("start %v:", c)
		next := s.Run(c.y)
		log.Printf("next %s -> %#v", c.y, next)
		if next == nil {
			if c.match {
				t.Fatalf("sad %v wanted match", c)
			}
		} else {
			if !c.match {
				t.Fatalf("sad %v wanted nomatch", c)
			}
		}
	}
}
