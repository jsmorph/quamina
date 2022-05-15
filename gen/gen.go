package gen

// It's so tempting to get carried away with interfaces for general
// probability distributions.  I WILL resist this time.  I WILL.

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"quamina/pruner"
)

func main() {
	spec := &Value{
		Map:    0.5,
		Array:  0.5,
		Int:    0.4,
		String: 0.6,
		Strings: String{
			Length: Int{
				Min: 5,
				Max: 30,
			},
		},
		Arrays: Array{
			Length: Int{
				Min: 1,
				Max: 5,
			},
		},
		Ints: Int{
			Min: -100,
			Max: 100,
		},
		Maps: Map{
			NumProperties: Int{
				Min: 2,
				Max: 5,
			},
			Properties: String{
				Length: Int{
					Min: 3,
					Max: 20,
				},
			},
		},
		Decays: Decays{
			Map:   0.5,
			Array: 0.8,
		},
	}

	pruner := &Pruner{
		Map:   0.5,
		Array: 0.5,
	}

	var (
		from      = flag.Int("from", 10, "starting number of iterations")
		to        = flag.Int("to", 20, "ending number of iterations")
		step      = flag.Int("step", 10, "step")
		repeat    = flag.Int("repeat", 3, "repeats")
		core      = flag.Bool("core", false, "use core matcher")
		seed      = flag.Int64("seed", time.Now().UnixNano(), "rand.Seed")
		noHeader  = flag.Bool("no-header", false, "no header lines")
		noRebuild = flag.Bool("no-rebuild", false, "no rebuilding")
	)

	flag.Float64Var(&spec.Decays.Map, "map-decay", spec.Decays.Map, "map decay")
	flag.Float64Var(&spec.Decays.Array, "array-decay", spec.Decays.Array, "array decay")
	flag.Float64Var(&pruner.Map, "prune-map", pruner.Map, "prune map rate")
	flag.Float64Var(&pruner.Array, "prune-array", pruner.Array, "prune array rate")

	flag.Parse()

	rand.Seed(*seed)

	if !*noHeader {
		fmt.Printf("matcher,op,impl,n,round,secs\n")
		fmt.Printf("data,kind,impl,n,round,value\n")
		fmt.Printf("dim,kind,impl,n,round,%s\n", DimsCSVHeader)
	}

	for i := *from; i <= *to; i += *step {
		for j := 0; j < *repeat; j++ {
			runtime.GC()
			many(i, j, *core, spec, pruner, *noRebuild)
		}
	}
}

type Int struct {
	Min, Max int
}

func (s *Int) Sample() int {
	return rand.Intn(s.Max-s.Min) + s.Min
}

type Float struct {
	Min, Max float64
}

func (s *Float) Sample() float64 {
	return 0
}

type Map struct {
	NumProperties Int
	Properties    String
}

type Decays struct {
	Map   float64
	Array float64
}

func (d *Decays) Decay(r *Decays) *Decays {
	return &Decays{
		Map:   d.Map * r.Map,
		Array: d.Array * r.Array,
	}
}

// var bab = (func() babble.Babbler {
// 	b := babble.NewBabbler()
// 	b.Count = 1
// 	return b
// })()

func (s *Map) Sample(v *Value, decays *Decays) map[string]interface{} {
	n := s.NumProperties.Sample()
	acc := make(map[string]interface{}, n)
	for i := 0; i < n; i++ {
		p := s.Properties.Sample()
		// p := bab.Babble()
		acc[p] = v.Sample(decays)
	}
	return acc
}

type Array struct {
	Length Int
}

func (s *Array) Sample(v *Value, decays *Decays) []interface{} {
	n := s.Length.Sample()
	acc := make([]interface{}, n)
	for i := 0; i < n; i++ {
		acc[i] = v.Sample(decays)
	}
	return acc
}

type Value struct {
	Map, Array, Int, Float, Bool, Null, String float64

	Maps    Map
	Strings String
	Ints    Int
	Arrays  Array
	Decays  Decays
}

type String struct {
	Length Int
}

type Char struct{}

func (s *Char) Sample() byte {
	chars := "abcdefghijklmnopqrstuvwxyz"
	i := rand.Intn(len(chars))
	return chars[i]
}

var char = &Char{}

func (s *String) Sample() string {
	n := s.Length.Sample()
	acc := make([]byte, n)
	for i := range acc {
		acc[i] = char.Sample()
	}
	return string(acc)
}

func (s *Value) Sample(decays *Decays) interface{} {
	if decays == nil {
		decays = &Decays{1, 1}
	}
	x := rand.Float64()
	f := s.Map * decays.Map
	if x < f {
		return s.Maps.Sample(s, decays.Decay(&s.Decays))
	}
	f += s.Array * decays.Array
	if x < f {
		return s.Arrays.Sample(s, decays.Decay(&s.Decays))
	}
	f += s.Int
	if x < f {
		return s.Ints.Sample()
	}

	return s.Strings.Sample()
}

type Pruner struct {
	Map   float64
	Array float64
}

func (p *Pruner) Prune(x interface{}) interface{} {
	switch vv := x.(type) {
	case map[string]interface{}:
		acc := make(map[string]interface{})
		for k, v := range vv {
			if rand.Float64() < p.Map {
				continue
			}
			acc[k] = p.Prune(v)
		}
		if len(acc) == 0 {
			for k, v := range vv {
				acc[k] = p.Prune(v)
				break
			}
		}
		return acc
	case []interface{}:
		acc := make([]interface{}, 0, len(vv))
		for _, v := range vv {
			if rand.Float64() < p.Array {
				continue
			}
			acc = append(acc, p.Prune(v))
		}
		return acc
	default:
		return x
	}
}

func Arrayify(x interface{}) interface{} {
	switch vv := x.(type) {
	case map[string]interface{}:
		acc := make(map[string]interface{})
		for k, v := range vv {
			acc[k] = Arrayify(v)
		}
		return acc
	case []interface{}:
		if len(vv) == 0 {
			return nil
		}
		// If there's a map in here, just return it.
		for _, v := range vv {
			if m, is := v.(map[string]interface{}); is {
				return Arrayify(m)
			}
		}
		// If there's an array in here, ignore it.
		// Return a subset of the array of atoms.
		atomics := make([]interface{}, 0, len(vv))
		for _, v := range vv {
			if _, is := v.([]interface{}); is {
				continue
			}
			atomics = append(atomics, v)
		}
		if 0 == len(atomics) {
			return nil
		}
		want := rand.Intn(len(atomics)) + 1
		acc := make([]interface{}, 0, want)
		for _, i := range rand.Perm(len(vv)) {
			v := vv[i]
			acc = append(acc, v)
			if len(acc) == want {
				break
			}
		}

		return acc
	default:
		return []interface{}{vv}
	}
}

func many(iters, round int, core bool, s *Value, p *Pruner, noRebuild bool) {

	var (
		m        PatternIndex
		events   = make(map[int]string, iters)
		patterns = make(map[int]string, iters)
		impl     string
	)

	if core {
		m = NewCoreMatcher()
		impl = "core"
	} else {
		m = pruner.NewMatcher(nil)
		m.(*pruner.Matcher).DisableRebuild()
		impl = "pruner"
	}

	i := 0
	for {

		event := s.Sample(nil)
		eventjs, err := json.Marshal(&event)
		if err != nil {
			log.Fatal(err)
		}
		if _, is := event.(map[string]interface{}); !is {
			// log.Printf("not attempting a %T event: %s", event, eventjs)
			continue
		}
		pruned := p.Prune(event)
		// prunedjs, err := json.MarshalIndent(&pruned, "", "  ")
		// if err != nil {
		// 	log.Fatal(err)
		// }

		pattern := Arrayify(pruned)
		patternjs, err := json.Marshal(&pattern)
		if err != nil {
			log.Fatal(err)
		}

		if m, is := pattern.(map[string]interface{}); is {
			if len(m) == 0 {
				continue
			}
		}

		if err := m.AddPattern(i, string(patternjs)); err != nil {
			// log.Printf("sad pattern %d %s: %v (pruned: %s)", i, patternjs, err, prunedjs)
			continue
		}

		events[i] = string(eventjs)
		patterns[i] = string(patternjs)
		if true {
			fmt.Printf("data,event,%s,%d,%d,%s\n", impl, i, round, eventjs)
			fmt.Printf("data,pattern,%s,%d,%d,%s\n", impl, i, round, eventjs)
		}
		fmt.Printf("dim,event,%s,%d,%d,%s\n", impl, i, round, ComputeDims(string(eventjs)).CSV())
		fmt.Printf("dim,pattern,%s,%d,%d,%s\n", impl, i, round, ComputeDims(string(patternjs)).CSV())

		i++
		if i == iters {
			break
		}
	}

	var passed, failed int

	then := time.Now()

EVENTS:
	for i, event := range events {
		got, err := m.MatchesForJSONEvent([]byte(event))
		if err != nil {
			// log.Printf("sad event %s: %v", event, err)
			continue
		}

		for _, j := range got {
			if i == j {
				// log.Printf("victory %d: %v", i, got)
				passed++
				continue EVENTS
			}
		}

		log.Printf("failed %d: %v", i, got)
		log.Printf("event %s", event)
		log.Printf("pattern %s", patterns[i])
		failed++
	}

	log.Printf("passed: %d, failed; %d", passed, failed)

	fmt.Printf("matcher,match,%s,%d,%d,%f\n", impl, iters, round, time.Now().Sub(then).Seconds())

	if !noRebuild {
		runtime.GC()
		then = time.Now()
		m.Rebuild(true)
		fmt.Printf("matcher,rebuild,%s,%d,%d,%f\n", impl, iters, round, time.Now().Sub(then).Seconds())
	}

}
