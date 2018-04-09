package ratelimit

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrNoSlotsAvailable ...
	ErrNoSlotsAvailable = errors.New("no slots available")
)

// Config ...
type Config struct {
	Rate  time.Duration
	Burst int
	Queue int
}

// Limiter ...
type Limiter struct {
	sync.Mutex

	Config Config

	lastSlotAt time.Time

	conns int
	queue *Queue

	value int
}

// NewLimiter ...
func NewLimiter(c Config) *Limiter {
	total := c.Burst + c.Queue

	return &Limiter{
		Config: c,
		queue:  NewQueue(),
		value:  total,
	}
}

// Release ...
func (l *Limiter) Release() {
	l.Lock()

	now := time.Now()

	if l.lastSlotAt.IsZero() {
		l.lastSlotAt = now
	}

	l.conns++

	l.Unlock()
}

// Take ...
func (l *Limiter) Take() error {
	calc := func() {
		if l.conns == 0 {
			if !l.lastSlotAt.IsZero() {
				// reset last slot at
				l.lastSlotAt = time.Time{}
			}
			return
		}

		now := time.Now()

		add := int(now.Sub(l.lastSlotAt) / l.Config.Rate)
		add = minInt(add, l.conns)
		if add == 0 {
			return
		}

		l.lastSlotAt = l.lastSlotAt.Add(time.Duration(add) * l.Config.Rate)

		l.conns -= add

		l.value += add

		l.queue.Free(add)
	}

	l.Lock()

	calc()

	if l.value == 0 {
		l.Unlock()
		return ErrNoSlotsAvailable
	}

	if l.value <= l.Config.Queue {
		c := l.queue.Add()

		if l.queue.Total() == 1 {
			go func() {
				for l.queue.Total() > 0 {
					time.Sleep(l.Config.Rate)

					l.Lock()
					calc()
					l.Unlock()
				}

			}()
		}

		l.value--

		l.Unlock()

		<-c

		return nil
	}

	l.value--

	l.Unlock()

	return nil
}

// Manager ...
type Manager struct {
	Config Config

	sync.Mutex
	buckets map[string]*Limiter
}

// Get ...
func (m *Manager) Get(b string) *Limiter {
	m.Lock()

	if m.buckets == nil {
		m.buckets = make(map[string]*Limiter)
	}

	l, ok := m.buckets[b]
	if !ok {
		l = NewLimiter(m.Config)
		m.buckets[b] = l
	}

	m.Unlock()

	return l
}

func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}
