package hops_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/ocpodariu/hops"
)

func ExampleCounter() {
	// Create a counter to track events from last 5 minutes
	c := hops.NewCounter(5, time.Minute)

	// Register events as they appear
	c.Observe()
	c.Observe()
	c.Observe()

	// Check number of registered events
	c.Value()
}

// TestCounterConcurrently is used to check for race conditions when reading
// and updating a counter at the same time.
//
// Run it with the race detector enabled:
//   $ go test -race -run TestCounterConcurrently
func TestCounterConcurrently(t *testing.T) {
	writer := func(c *hops.Counter, shutdown chan struct{}) {
		for {
			select {
			case <-shutdown:
				return
			default:
			}
			c.Observe()
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
		}
	}
	reader := func(c *hops.Counter, shutdown chan struct{}) {
		for {
			select {
			case <-shutdown:
				return
			default:
			}
			c.Value()
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		}
	}

	c := hops.NewCounter(5, time.Second)
	shutdown := make(chan struct{})

	// Start a couple of writers and readers
	for i := 0; i < 100; i++ {
		go writer(c, shutdown)
	}
	for i := 0; i < 50; i++ {
		go reader(c, shutdown)
	}

	// Let them run for a while
	time.Sleep(10 * time.Second)
	close(shutdown)
	time.Sleep(time.Second)
}
