package core_test

import (
	"sync"
	"testing"

	"github.com/dboslee/job-worker/pkg/core"
)

var want = []byte("test output")

func TestPubSub(t *testing.T) {
	stream := core.NewStream()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		ch := make(chan []byte, 1)
		stream.Subscribe(ch)

		go func() {
			defer wg.Done()
			got := <-ch
			if string(got) != string(want) {
				t.Errorf("pubsub want: %v got: %v", string(want), string(got))
			}
		}()
	}

	stream.Publish(want)
	wg.Wait()
}

// Test publish after unsubscribe.
func TestUnsubscribe(t *testing.T) {
	stream := core.NewStream()
	ch := make(chan []byte, 1)

	stream.Subscribe(ch)
	stream.Unsubscribe(ch)

	stream.Publish(want)
	if _, ok := <-ch; ok {
		t.Errorf("expected channel to be closed")
	}

}

func TestClose(t *testing.T) {
	stream := core.NewStream()
	ch := make(chan []byte, 1)
	stream.Subscribe(ch)
	stream.Close()
	if _, ok := <-ch; ok {
		t.Errorf("expected channel to be closed")
	}
}
