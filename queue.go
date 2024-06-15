package delayqueue

import (
	"container/heap"
	"context"
	"time"
)

type Queue[T any] struct {
	C <-chan T

	ctx       context.Context
	ch        chan T
	additions chan item[T]
	items     pqueue[T]
}

func New[T any](ctx context.Context, outBufferSize int) *Queue[T] {
	out := &Queue[T]{
		ctx:       ctx,
		ch:        make(chan T, outBufferSize),
		additions: make(chan item[T]),
		items:     make(pqueue[T], 0),
	}
	out.C = out.ch
	return out
}

func (q *Queue[T]) Run() {
	defer close(q.ch)

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	var nextDueAt time.Time

	for {
		select {
		case i := <-q.additions:
			heap.Push(&q.items, i)
			if nextDueAt.IsZero() || i.Due.Before(nextDueAt) {
				if !nextDueAt.IsZero() && !timer.Stop() {
					<-timer.C
				}
				nextDueAt = i.Due
				timer.Reset(time.Until(nextDueAt))
			}
		case <-timer.C:
			out := heap.Pop(&q.items).(item[T])
			if q.items.Len() > 0 {
				nextDueAt = q.items[0].Due
				timer.Reset(time.Until(nextDueAt))
			} else {
				nextDueAt = time.Time{}
			}
			q.ch <- out.V
		case <-q.ctx.Done():
			return
		}
	}
}

func (q *Queue[T]) Add(due time.Time, val T) error {
	select {
	case <-q.ctx.Done():
		return q.ctx.Err()
	case q.additions <- item[T]{Due: due, V: val}:
		return nil
	}
}

type item[T any] struct {
	Due time.Time
	V   T
}

type pqueue[T any] []item[T]

func (q pqueue[T]) Len() int           { return len(q) }
func (q pqueue[T]) Less(i, j int) bool { return q[i].Due.Before(q[j].Due) }
func (q pqueue[T]) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *pqueue[T]) Push(x any)        { *q = append(*q, x.(item[T])) }

func (q *pqueue[T]) Pop() any {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[0 : n-1]
	return x
}
