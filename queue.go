package delayqueue

import (
	"container/heap"
	"container/list"
	"context"
	"sync"
	"time"
)

type Queue[T any] struct {
	// Items added to the queue are sent to this channel as they become due.
	// If the number of items that will be added to the queue exceeds the buffer
	// size, this channel must be continously read in order to prevent blocking.
	C <-chan T

	ctx       context.Context
	ch        chan T
	additions chan item[T]
	items     pqueue[T]

	readyLock   sync.Mutex
	readySignal *sync.Cond
	ready       list.List
}

// Create a new Queue that will run until the provided context is cancelled.
// The second argument specifies the buffer size of C, the outgoing channel.
func New[T any](ctx context.Context, outBufferSize int) *Queue[T] {
	out := &Queue[T]{
		ctx:       ctx,
		ch:        make(chan T, outBufferSize),
		additions: make(chan item[T]),
		items:     make(pqueue[T], 0),
	}
	out.C = out.ch
	out.readySignal = sync.NewCond(&out.readyLock)
	go out.dispatch()
	go out.run()
	return out
}

func (q *Queue[T]) dispatch() {
	q.readyLock.Lock()
	for {
		if q.ctx.Err() != nil {
			return
		}
		for q.ready.Len() == 0 {
			q.readySignal.Wait()
		}
		i := q.ready.Remove(q.ready.Front()).(T)
		select {
		case q.ch <- i:
		case <-q.ctx.Done():
			return
		}
	}
}

// Run the queue, processing new additions and emitting existing items as
// they become due. run() will keep running until the context.Context passed
// to New() is cancelled, at which point the outgoing channel C is closed and
// the method will return.
func (q *Queue[T]) run() {
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
			q.readyLock.Lock()
			q.ready.PushBack(out.V)
			q.readyLock.Unlock()
			q.readySignal.Signal()
		case <-q.ctx.Done():
			return
		}
	}
}

// Add an item i to the queue, to be emitted at or after the given due time.
// Add() will return an error if the queue's context is cancelled before the
// operation completes; the returned error will be ctx.Err().
// Add() can be called safely from multiple goroutines.
func (q *Queue[T]) Add(due time.Time, i T) error {
	select {
	case <-q.ctx.Done():
		return q.ctx.Err()
	case q.additions <- item[T]{Due: due, V: i}:
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
