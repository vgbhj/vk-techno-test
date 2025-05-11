package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPublishSubscribeOrder(t *testing.T) {
	ps := NewSubPub()
	defer ps.Close(context.Background())

	var mu sync.Mutex
	received := []int{}

	_, err := ps.Subscribe("topic", func(msg interface{}) {
		mu.Lock()
		received = append(received, msg.(int))
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Subscribe error: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := ps.Publish("topic", i); err != nil {
			t.Fatalf("Publish error: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(received))
	}
	for i := 0; i < 5; i++ {
		if received[i] != i {
			t.Errorf("at index %d: expected %d, got %d", i, i, received[i])
		}
	}
}

func TestSlowSubscriberDoesNotBlock(t *testing.T) {
	ps := NewSubPub()
	defer ps.Close(context.Background())

	var fast, slow []int
	var wg sync.WaitGroup
	wg.Add(2)

	_, _ = ps.Subscribe("topic", func(msg interface{}) {
		time.Sleep(200 * time.Millisecond)
		slow = append(slow, msg.(int))
		wg.Done()
	})
	_, _ = ps.Subscribe("topic", func(msg interface{}) {
		fast = append(fast, msg.(int))
		wg.Done()
	})

	if err := ps.Publish("topic", 1); err != nil {
		t.Fatalf("Publish error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("handlers did not complete in time")
	}

	if len(fast) != 1 {
		t.Errorf("fast subscriber: expected 1, got %d", len(fast))
	}
	if len(slow) != 1 {
		t.Errorf("slow subscriber: expected 1, got %d", len(slow))
	}
}

func TestUnsubscribe(t *testing.T) {
	ps := NewSubPub()
	defer ps.Close(context.Background())

	var received []int
	sub, _ := ps.Subscribe("topic", func(msg interface{}) {
		received = append(received, msg.(int))
	})

	sub.Unsubscribe()
	ps.Publish("topic", 42)
	time.Sleep(50 * time.Millisecond)

	if len(received) != 0 {
		t.Errorf("expected 0 messages after unsubscribe, got %d", len(received))
	}
}

func TestCloseWaits(t *testing.T) {
	ps := NewSubPub()

	var processed bool
	_, _ = ps.Subscribe("topic", func(msg interface{}) {
		time.Sleep(100 * time.Millisecond)
		processed = true
	})

	ps.Publish("topic", "data")
	if err := ps.Close(context.Background()); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	if !processed {
		t.Error("expected message to be processed before Close returned")
	}
}

func TestCloseWithCancel(t *testing.T) {
	ps := NewSubPub()

	_, _ = ps.Subscribe("topic", func(msg interface{}) {
		time.Sleep(500 * time.Millisecond)
	})

	ps.Publish("topic", "data")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := ps.Close(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error on Close with canceled context, got nil")
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("Close did not return promptly after cancel; elapsed %v", elapsed)
	}
}
