package quamina

import "testing"

func TestUseStdExtension(t *testing.T) {
	t.Cleanup(func() {
		ThePredicateParser = nil
	})

	if err := UseStdExtension(); err != nil {
		t.Fatal(err)
	}

	q, err := New(WithPatternDeletion(true))
	if err != nil {
		t.Fatal(err)
	}

	pat := `{"likes":[{"numeric":["<",10]}]}`
	msg := `{"likes":5}`

	if err := AddExtendedPattern(q, 1, pat); err != nil {
		t.Fatal(err)
	}

	if err := AddExtendedPattern(q, 2, pat); err != nil {
		t.Fatal(err)
	}

	ids, err := q.MatchesForEvent([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatal(len(ids))
	}

	if err = q.DeletePatterns(1); err != nil {
		t.Fatal(err)
	}

	ids, err = q.MatchesForEvent([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Fatal(len(ids))
	}

}
