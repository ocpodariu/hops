## hops

hops is a hopping window counter that keeps track of how many events happened in the last N time units, with a hop size of 1 time unit. It's also safe to use concurrently by multiple readers and writers.

### Use case
The main problem that it solves and the reason I've built it is to count a large number of events over a period of time, while using a constant amount of memory regardless of how many events occur.

This counter implementation is right for you if the sentences below apply to your use case:
- You want to keep track of how many events happened in the last N time units (e.g. last 5 minutes).
- You want to count a large number of events (thousands, hundreds of thousands, millions etc.).
- You care only about the number of events. The details of each event are irrelevant to you.

As a sidenote, a simple integer variable that is incremented for each event is not enough to solve this problem. Old events must be expired once they are outside of the time window of interest. Since the integer variable has no idea of the timestamps when it was incremented, it doesn't know by what amount to decrement itself after each unit of time passes. Therefore, by itself, an integer variable is unable to keep track of a moving count.

### How to get
```
go get github.com/ocpodariu/hops
```

### Example - Basic usage
```
package main

import (
    "fmt"
    "time"

    "github.com/ocpodariu/hops"
)

func main() {
	// Create a counter to track events from last 5 minutes
	c := hops.NewCounter(5, time.Minute)

	// Register events as they appear
	c.Observe()
	c.Observe()
	c.Observe()

	// Check number of registered events
	fmt.Println(c.Value())
}
```

### Example - Count HTTP requests
Keep track of how many requests you've handled in the last 5 seconds.

```
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/ocpodariu/hops"
)

func main() {
	c := hops.NewCounter(5, time.Second)

	http.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
		c.Observe()

		// Do some work
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

		fmt.Fprint(w, c.Value())
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```
