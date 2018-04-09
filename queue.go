package ratelimit

// Queue ...
type Queue struct {
	c     chan struct{}
	total int
}

// NewQueue ...
func NewQueue() *Queue {
	return &Queue{
		c: make(chan struct{}),
	}
}

// Free ...
func (q *Queue) Free(n int) {
	for ; n > 0 && q.total > 0; n-- {
		q.total--
		q.c <- struct{}{}
	}
}

// Add ...
func (q *Queue) Add() <-chan struct{} {
	q.total++

	return q.c
}

// Total ...
func (q *Queue) Total() int {
	return q.total
}
