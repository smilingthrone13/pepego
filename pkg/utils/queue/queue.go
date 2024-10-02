package queue

import (
	"slices"
)

func NewQueue(queueSize int) *Queue {
	return &Queue{
		s:   make([]string, 0, queueSize),
		cap: queueSize,
	}
}

type Queue struct {
	s    []string
	cap  int
	head int
	size int
}

func (q *Queue) Add(v string) {
	if q.size < q.cap {
		q.s = append(q.s, v)
		q.size++
	} else {
		q.s[q.head] = v
		q.head = (q.head + 1) % q.cap
	}
}

func (q *Queue) GetAll() []string {
	return q.s
}

func (q *Queue) Contains(v string) bool {
	return slices.Contains(q.s, v)
}
