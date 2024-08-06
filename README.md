# delayqueue

[![](https://godoc.org/github.com/jaz303/delayqueue?status.svg)](http://godoc.org/github.com/jaz303/delayqueue)

This is a simple delay queue that sends each added value to a channel after a specified delay has elapsed. Compared to the naive approach of spawning one goroutine per item, this implementation uses a constant two goroutines, a single auxiliary timer, and a priority queue, so should therefore comfortably handle a large number of items.

## Install

```shell
go get github.com/jaz303/delayqueue
```

## Usage

```go
// Create a new delay queue of ints that will run until
// the provided context is cancelled. The second argument
// is the buffer size of the outgoing channel.
queue := delayqueue.New[int](context.Background(), 0)

go func() {
    // Read items from the queue
    for i := range queue.C {
        log.Printf("Read from queue: %+v", i)
    }
}()

now := time.Now()

// Add items to the queue
queue.Add(now.Add(100 * time.Millisecond), 2)
queue.Add(now.Add(500 * time.Millisecond), 3)
queue.Add(now, 1)
```

# Benchmarks

`benchmark_test.go` compares `delayqueue` to a goroutine per item approach:

```
BenchmarkDelayQueue-4   	       1	8313947735 ns/op	320948480 B/op	10001373 allocs/op
BenchmarkNaive-4        	       1	10157876788 ns/op	3296331856 B/op	19999860 allocs/op
```

`delayqueue` is about 20% faster, performs half as many allocations, and uses a whopping **10 times** less RAM.
