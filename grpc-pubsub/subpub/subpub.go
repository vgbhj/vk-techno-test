package subpub

import (
	"context"
	"errors"
	"sync"
)

// MessageHandler processes messages for subscribers.
type MessageHandler func(msg interface{})

// Subscription allows unsubscribing from a subject.
type Subscription interface {
	Unsubscribe()
}

// SubPub defines the pub-sub interface.
type SubPub interface {
	// Subscribe creates an asynchronous subscriber on the given subject.
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg interface{}) error
	// Close will shutdown sub-pub system.
	// May be blocked by data delivery until the context is canceled.
	Close(ctx context.Context) error
}

// NewSubPub creates a new SubPub instance.
func NewSubPub() SubPub {
	return &pubsub{
		subs: make(map[string][]*subscription),
	}
}

type pubsub struct {
	mu     sync.RWMutex
	subs   map[string][]*subscription
	wg     sync.WaitGroup
	closed bool
}

type subscription struct {
	subj   string
	cb     MessageHandler
	mu     sync.Mutex
	cond   *sync.Cond
	queue  []interface{}
	closed bool
	ps     *pubsub
}

func (ps *pubsub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	ps.mu.Lock()
	if ps.closed {
		ps.mu.Unlock()
		return nil, errors.New("subpub is closed")
	}
	s := &subscription{subj: subject, cb: cb, ps: ps}
	s.cond = sync.NewCond(&s.mu)
	ps.subs[subject] = append(ps.subs[subject], s)
	ps.wg.Add(1)
	ps.mu.Unlock()

	go s.loop()
	return s, nil
}

func (s *subscription) loop() {
	defer s.ps.wg.Done()
	for {
		s.mu.Lock()
		for len(s.queue) == 0 && !s.closed {
			s.cond.Wait()
		}
		if len(s.queue) == 0 && s.closed {
			s.mu.Unlock()
			return
		}
		msg := s.queue[0]
		s.queue = s.queue[1:]
		s.mu.Unlock()
		// deliver
		s.cb(msg)
	}
}

func (s *subscription) Unsubscribe() {
	s.ps.mu.Lock()
	subs := s.ps.subs[s.subj]
	for i, sub := range subs {
		if sub == s {
			s.ps.subs[s.subj] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(s.ps.subs[s.subj]) == 0 {
		delete(s.ps.subs, s.subj)
	}
	s.mu.Lock()
	s.closed = true
	s.cond.Signal()
	s.mu.Unlock()

	s.ps.mu.Unlock()
}

func (ps *pubsub) Publish(subject string, msg interface{}) error {
	ps.mu.RLock()
	if ps.closed {
		ps.mu.RUnlock()
		return errors.New("subpub is closed")
	}
	subs := append([]*subscription(nil), ps.subs[subject]...)
	ps.mu.RUnlock()

	for _, s := range subs {
		s.mu.Lock()
		s.queue = append(s.queue, msg)
		s.cond.Signal()
		s.mu.Unlock()
	}
	return nil
}

func (ps *pubsub) Close(ctx context.Context) error {
	ps.mu.Lock()
	if ps.closed {
		ps.mu.Unlock()
		return nil
	}
	ps.closed = true
	allSubs := make([]*subscription, 0)
	for _, sl := range ps.subs {
		allSubs = append(allSubs, sl...)
	}
	ps.subs = make(map[string][]*subscription)
	ps.mu.Unlock()

	for _, s := range allSubs {
		s.mu.Lock()
		s.closed = true
		s.cond.Signal()
		s.mu.Unlock()
	}

	done := make(chan struct{})
	go func() {
		ps.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
