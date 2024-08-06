package delayqueue

import (
	"context"
	"sync"
	"testing"
	"time"
)

const count = 5_000_000

func BenchmarkDelayQueue(b *testing.B) {
	queue := New[int](context.Background(), 0)

	now := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i := 0; i < count; i++ {
			<-queue.C
		}
		wg.Done()
	}()

	for i := 0; i < count; i++ {
		queue.Add(now, i)
	}

	wg.Wait()
}

func BenchmarkNaive(b *testing.B) {
	results := make(chan int)

	wg := sync.WaitGroup{}
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			results <- i
			wg.Done()
		}(i)
	}

	for i := 0; i < count; i++ {
		<-results
	}

	wg.Wait()
}
