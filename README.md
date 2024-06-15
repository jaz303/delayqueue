# delayqueue

This is a simple delay queue that sends each added value to a channel after a specified delay has elapsed. Compared to the naive approach of spawning one goroutine per item, this implementation uses only a single main goroutine, a single auxiliary timer, and a priority queue, so should therefore comfortably handle a large number of items.

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