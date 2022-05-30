package async

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	m := NewMatcher()
	m.Logging = true
	m.RebuildReports = make(chan RebuildReport)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.Run(ctx, nil); err != nil {
			t.Fatal(err)
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case r := <-m.RebuildReports:
				log.Printf("rebuild report %#v", r)
			}
		}
	}()

	if err := m.AddPattern(ctx, 1, `{"a":[1]}`); err != nil {
		t.Fatal(err)
	}

LOOP:
	for i := 0; true; i++ {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}
		got, err := m.MatchesForJSONEvent(ctx, []byte(`{"a":1}`))
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("got %#v", got)

		if i == 2 {
			if err = m.DeletePattern(ctx, 1); err != nil {
				t.Fatal(err)
			}
		} else {
			time.Sleep(600 * time.Millisecond)
		}
	}

	wg.Wait()
}
