package gen

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	quamina "github.com/timbray/quamina/core"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

func TestMatcher(t *testing.T) {
	qcheck := func(e, p string) (bool, bool) {
		var x interface{}
		if err := json.Unmarshal([]byte(e), &x); err != nil {
			t.Fatal("badjson", e)
		}
		if err := json.Unmarshal([]byte(p), &x); err != nil {
			t.Fatal("badjson", p)
		}
		m := quamina.NewCoreMatcher()
		if err := m.AddPattern(1, p); err != nil {
			return false, false
		}
		got, err := m.MatchesForJSONEvent([]byte(e))

		if err != nil {
			log.Printf("aws error %v", err)
			return false, false
		}
		return len(got) == 1, true
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("unable to load AWS SDK config, %v", err)
	}

	eb := eventbridge.NewFromConfig(cfg)

	awscheck := func(e, p string) (bool, bool) {
		matches, err := EventBridgeMatches(ctx, eb, p, e)
		if err != nil {
			log.Printf("aws error %v", err)
			return false, false
		}
		return matches, true
	}

	check := func(e, p string) {
		q1, q2 := qcheck(e, p)
		a1, a2 := awscheck(e, p)
		if q1 != a1 || q2 != a2 {
			t.Fatalf("%v/%v %v/%v %s %s\n", q1, q2, a1, a2, p, e)
		}
	}

	type PE struct {
		Pattern string
		Event   string
	}

	for _, pe := range []PE{
		{`{"a":[{"b":"c"}]}`, `{"a":{"b":["c"]}}`},
		{`{"a":"b"}`, `{}`},
	} {
		check(pe.Pattern, pe.Event)
	}
}
