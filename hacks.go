package quamina

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

var (
	EnableHacks       = true
	FloatNearMaxDelta = 1e-08
)

func hasHack(s string) (*Hack, bool) {
	if !EnableHacks {
		return nil, false
	}

	// "{...}" (with quotes in string)

	n := len(s)
	if n < 5 {
		return nil, false
	}
	if s[0] != '"' {
		return nil, false
	}
	if s[1] != '{' {
		return nil, false
	}
	if s[n-2] != '}' {
		return nil, false
	}
	if s[n-1] != '"' {
		return nil, false
	}
	js := s[1 : n-1]

	var h Hack
	if err := json.Unmarshal([]byte(js), &h); err != nil {
		return nil, false
	}

	if err := h.Compile(); err != nil {
		return nil, false
	}

	return &h, true
}

type Hack struct {
	Numeric *NumericHack `json:"numeric,omitempty"`
}

func (h *Hack) Matches(val []byte) bool {
	if !EnableHacks {
		return false
	}
	if h.Numeric != nil {
		return h.Numeric.Matches(val)
	}
	return false
}

func (h *Hack) Compile() error {
	if !EnableHacks {
		return nil
	}
	if h.Numeric != nil {
		if err := h.Numeric.Compile(); err != nil {
			return err
		}
	}
	return nil
}

type NumericHack []any

func (h *NumericHack) Compile() error {
	cs, err := CompileNumericConstraints(*h)
	if err != nil {
		return err
	}

	// Please look away.
	(*h)[0] = cs
	return nil
}

func (h *NumericHack) Matches(val []byte) bool {
	x, err := strconv.ParseFloat(string(val), 63)
	if err != nil {
		return false
	}
	cs, is := (*h)[0].(NumericConstraints)
	if !is {
		return false
	}
	matches, err := cs.Matches(x)
	if err != nil {
		return false
	}

	return matches
}

func HackPat(pat string) (string, error) {
	var x any
	if err := json.Unmarshal([]byte(pat), &x); err != nil {
		return "", err
	}
	y, err := UseHacks(x)
	if err != nil {
		return "", err
	}
	js, err := niceJSON(y)
	if err != nil {
		return "", err
	}
	return string(js), nil
}

func UseHacks(x any) (any, error) {
	if !EnableHacks {
		return x, nil
	}

	switch vv := x.(type) {
	case []any:
		acc := make([]any, 0, len(vv))
		if len(vv) == 1 {
			switch vv := vv[0].(type) {
			case map[string]any:
				if len(vv) == 1 {
					if numeric, has := vv["numeric"]; has {
						x := map[string]any{
							"numeric": numeric,
						}
						js, err := niceJSON(x)
						if err != nil {
							return nil, err
						}
						return append(acc, string(js)), nil
					}
				}
			}
		}

		for _, y := range vv {
			acc = append(acc, y)
		}

		return acc, nil
	case map[string]any:
		acc := make(map[string]any, len(vv))
		for p, v := range vv {
			y, err := UseHacks(v)
			if err != nil {
				return nil, err
			}
			acc[p] = y
		}
		return acc, nil
	default:
		return x, nil
	}
}

func niceJSON(x interface{}) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(&x)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(buffer.Bytes())), nil
}
