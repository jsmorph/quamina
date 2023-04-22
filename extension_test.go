package quamina

import (
	"log"
	"testing"
)

func TestExtension0(t *testing.T) {
	q, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = q.AddPattern(1, `{"likes":[{"extension":"tacos"}]}`)
	if err != nil {
		t.Fatal(err)
	}

	{
		matches, err := q.MatchesForEvent([]byte(`{"likes":"tacos"}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) == 0 {
			t.Fatal("expected matches")
		}
	}

	{
		matches, err := q.MatchesForEvent([]byte(`{"likes":"some tacos"}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 0 {
			t.Fatal(len(matches))
		}
	}
}

func TestExtension1(t *testing.T) {

	ThePredicateParser = func(spec []byte) (Predicate, error) {
		log.Printf("debug ThePredicateParser parsing %s", spec)
		return func(bs []byte) bool {
			log.Printf("debug Predicate considering %s (len: %d)", bs, len(bs))
			return len(bs) == 7
		}, nil
	}

	q, err := New()
	if err != nil {
		t.Fatal(err)
	}

	err = q.AddPattern(1, `{"likes":[{"extension":{"loves":"queso"}}]}`)
	if err != nil {
		t.Fatal(err)
	}

	{
		matches, err := q.MatchesForEvent([]byte(`{"likes":"tacos"}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) == 0 {
			t.Fatal("expected matches")
		}
	}

	{
		matches, err := q.MatchesForEvent([]byte(`{"likes":"some tacos"}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 0 {
			t.Fatal(len(matches))
		}
	}
}
