## hops

hops is a hopping window counter that keeps track of how many events happened in the last N time units, with a hop size of 1 time unit. It's also safe to use concurrently by multiple readers and writers.

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
