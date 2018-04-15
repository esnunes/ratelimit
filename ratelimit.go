package ratelimit

import (
	"errors"
	"sync"
	"time"

	"github.com/esnunes/ratelimit/pkg/util"
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

	config Config

	currentToExpire time.Time

	queue *Queue

	toExpire int
	slots    int
}

// NewLimiter ...
func NewLimiter(c Config) *Limiter {
	return &Limiter{
		config: c,
		queue:  NewQueue(),
		slots:  c.Burst + c.Queue,
	}
}

// Use ...
func (l *Limiter) Use() {
	l.Lock()

	if l.currentToExpire.IsZero() {
		// first element to expire
		l.currentToExpire = time.Now()
	}

	l.toExpire++

	l.Unlock()
}

// Reserve ...
func (l *Limiter) Reserve() error {
	l.Lock()

	l.checkExpired()

	if l.slots == 0 {
		l.Unlock()
		return ErrNoSlotsAvailable
	}

	l.slots--

	if l.slots < l.config.Queue {
		c := l.queue.Add()

		if l.queue.Total() == 1 {
			go l.checkQueueExpiredLoop()
		}

		l.Unlock()

		<-c

		return nil
	}

	l.Unlock()

	return nil
}

func (l *Limiter) checkExpired() {
	if l.toExpire == 0 {
		return
	}

	expired := int(time.Now().Sub(l.currentToExpire) / l.config.Rate)
	expired = util.MinInt(expired, l.toExpire)
	if expired == 0 {
		return
	}

	l.currentToExpire = l.currentToExpire.Add(time.Duration(expired) * l.config.Rate)

	l.toExpire -= expired
	if l.toExpire == 0 {
		// reset currentToExpire value
		l.currentToExpire = time.Time{}
	}

	l.slots += expired

	l.queue.Free(expired)
}

func (l *Limiter) checkQueueExpiredLoop() {
	for l.queue.Total() > 0 {
		time.Sleep(l.config.Rate)

		l.Lock()
		l.checkExpired()
		l.Unlock()
	}
}

// Manager ...
type Manager struct {
	Config Config

	sync.Mutex
	limiters map[string]*Limiter
}

// Get ...
func (m *Manager) Get(b string) *Limiter {
	m.Lock()

	if m.limiters == nil {
		m.limiters = make(map[string]*Limiter)
	}

	l, ok := m.limiters[b]
	if !ok {
		l = NewLimiter(m.Config)
		m.limiters[b] = l
	}

	m.Unlock()

	return l
}
