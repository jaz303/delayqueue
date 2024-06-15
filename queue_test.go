package delayqueue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestElementsYieldedInTimeOrder(t *testing.T) {
	q := New[int](context.Background(), 3)
	q.Add(time.Now().Add(100*time.Millisecond), 2)
	q.Add(time.Now().Add(200*time.Millisecond), 3)
	q.Add(time.Now().Add(50*time.Millisecond), 1)

	collect := []int{}
	for i := 0; i < 3; i++ {
		collect = append(collect, <-q.C)
	}

	assert.Equal(t, 1, collect[0])
	assert.Equal(t, 2, collect[1])
	assert.Equal(t, 3, collect[2])
}

const epsilon = 5 * time.Millisecond

func TestElementsYieldedAtCorrectTime(t *testing.T) {
	now := time.Now()

	vals := []int{1, 2, 3, 4, 5}
	dues := []time.Time{
		now.Add(100 * time.Millisecond),
		now.Add(300 * time.Millisecond),
		now.Add(600 * time.Millisecond),
		now.Add(650 * time.Millisecond),
		now.Add(1 * time.Second),
	}

	q := New[int](context.Background(), 5)
	for i := range vals {
		q.Add(dues[i], vals[i])
	}

	for i := range vals {
		v := <-q.C
		assert.Equal(t, vals[i], v)
		dt := time.Since(dues[i])
		if dt < 0 {
			dt = -dt
		}
		assert.True(t, dt < epsilon)
	}
}
