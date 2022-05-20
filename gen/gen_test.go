package gen

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDims(t *testing.T) {
	s := `{"drhtuaiz":-39,"xiylvbnxhkpmmidcp":{"loeohqs":[92,-80,"wyuolcwcbdixvddtwgqoyukjrvsk",-49],"lvnd":[["ffjugczdbuciuubbqbokrzuow"],[-94,"wxyftndvcsmjrvp",0,-35],[[92]]],"nafaouhyiyawrpjam":-2}}`

	fmt.Printf("%#v\n", ComputeDims(s))

}

func TestGen(t *testing.T) {
	s := Value{
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

	x := s.Sample(nil)
	js, err := json.MarshalIndent(&x, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("event %s\n", js)

	pruner := &Pruner{
		Map:   0.5,
		Array: 0.5,
	}
	x = pruner.Prune(x)
	js, err = json.MarshalIndent(&x, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("pruned %s\n", js)

	x = Arrayify(x)
	js, err = json.MarshalIndent(&x, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("pattern %s\n", js)

}
