package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jsmorph/quamina"
)

func main() {
	var (
		clean    = flag.Bool("clean", false, "do not used hacked Quamina")
		patterns = flag.String("patterns", "", "patterns")
		pattern  = flag.String("pattern", "", "pattern")
		mode     = flag.String("mode", "filter", "filter|boolean|invert")
	)

	flag.Parse()

	if !*clean {
		if err := quamina.UseStdExtension(); err != nil {
			panic(err)
		}
	}

	q, err := quamina.New()
	if err != nil {
		panic(err)
	}

	if *pattern != "" {
		if *patterns != "" {
			panic("given one or the other of -patterns and -pattern")
		}
		*patterns = `[` + *pattern + `]`
	}

	if *patterns == "" {
		panic("give either -patterns or -pattern")
	}

	var pats []any
	if err := json.Unmarshal([]byte(*patterns), &pats); err != nil {
		panic(err)
	}

	render := func(x any) string {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(&x)
		if err != nil {
			panic(err)
		}
		return strings.TrimSpace(string(buffer.Bytes()))
	}

	for _, pat := range pats {
		js := render(pat)
		if err := quamina.AddExtendedPattern(q, 0, js); err != nil {
			panic(err)
		}
	}

	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		xs, err := q.MatchesForEvent([]byte(line))
		if err != nil {
			panic(err)
		}
		matches := 0 < len(xs)
		switch *mode {
		case "filter":
			if matches {
				fmt.Printf("%v\n", line)
			}
		case "invert":
			if !matches {
				fmt.Printf("%v\n", line)
			}
		case "boolean":
			fmt.Printf("%v", matches)
		default:
			panic(fmt.Errorf("unknown mode: %s", *mode))
		}
	}

	if err := in.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

}
