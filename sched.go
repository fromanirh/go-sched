package sched

import (
	pq "github.com/cenkalti/gopqueue"
	"runtime"
	"sync"
	"time"
)

type Event struct {
	ts     time.Time
	action func()
}

func (e *Event) Less(other interface{}) bool {
	return e.ts.Before(other.(*Event).ts)
}

type Scheduler struct {
	// if true, executes each event in a separate goroutine.
	useGoroutines bool
	now           func() time.Time
	sleep         func(d time.Duration)
	queue         *pq.Queue
	lock          sync.RWMutex
}

// NewWithHooks creates a new scheduler replacing time functions, and return it's pointer.
func NewWithHooks(useGoroutines bool, now func() time.Time, sleep func(d time.Duration)) *Scheduler {
	return &Scheduler{
		useGoroutines: useGoroutines,
		now:           now,
		sleep:         sleep,
		queue:         pq.New(0),
	}
}

// News creates a new scheduler, and return it's pointer.
func New() *Scheduler {
	return &Scheduler{
		useGoroutines: true,
		now:           time.Now,
		sleep:         time.Sleep,
		queue:         pq.New(0),
	}
}

// EnterAbs adds a new event to the queue at an absolute time.
func (s *Scheduler) EnterAbs(ts time.Time, action func()) Event {
	s.lock.Lock()
	defer s.lock.Unlock()

	event := Event{ts, action}
	s.queue.Enqueue(&event)
	return event
}

// Enter adds an new event to the queue to run after delay.
func (s *Scheduler) Enter(delay time.Duration, action func()) Event {
	diff := s.now().Add(delay)
	return s.EnterAbs(diff, action)
}

// Empty returns true if there is not events in the queue.
func (s *Scheduler) Empty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.queue.IsEmpty()
}

// Len returns the number of items in the scheduler's event queue.
func (s *Scheduler) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.queue.Len()
}

// Run executes events until the queue is empty.
func (s *Scheduler) Run() {
	var delay bool

	for {
		s.lock.Lock()
		min := s.queue.Peek()
		if min == nil {
			s.lock.Unlock()
			break
		}

		event := min.(*Event)
		now := s.now()

		if event.ts.After(now) {
			delay = true
		} else {
			delay = false
			s.queue.Dequeue()
		}
		s.lock.Unlock()

		if delay {
			s.sleep(event.ts.Sub(now))
		} else if s.useGoroutines {
			go event.action()
			runtime.Gosched() // Don't know if this is required
		} else {
			event.action()
		}
	}
}
