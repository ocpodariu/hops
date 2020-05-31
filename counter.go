package hops

import (
	"sync"
	"sync/atomic"
	"time"
)

// Counter uses a hopping window to keep track of how many events happened
// in the last W time units, with a hop size of 1 time unit.
//
// It's safe to use this counter concurrently.
type Counter struct {
	// Number of events that happen in the current time unit.
	// Use only atomic operations to read and write to this field.
	crtCount uint32

	// Guards prevCounts and windowStart
	mu sync.RWMutex

	// Number of events that happened in each of the last (W-1) time units.
	// prevCounts[i] = number of events that happened (W-1-i) time units ago
	//
	// Example for a 4-minute window:
	//   prevCounts[0] = total events that happened 3 minutes ago
	//   prevCounts[1] = total events that happened 2 minutes ago
	//   prevCounts[2] = total events that happened 1 minute ago
	prevCounts []uint32

	windowStart time.Time

	WindowSize time.Duration
	Unit       time.Duration
}

// NewCounter creates a new counter with the given window size and time unit.
//
// For example, NewCounter(5, time.Minute) creates a counter that keeps track
// of how many events happened in the last 5 minutes.
func NewCounter(windowSize int, timeUnit time.Duration) *Counter {
	// Initialize the window such that its end is on the current time unit.
	//
	// For example, if you create a 5-minute window at 15:21:43, then the
	// window start will be at 15:17 and the window end at 15:21. The window
	// covers events between 15:17:00 and 15:21:59.
	windowStart := time.Now().Truncate(timeUnit).Add(timeUnit)
	windowStart = windowStart.Add(-1 * time.Duration(windowSize) * timeUnit)

	return &Counter{
		crtCount:    0,
		prevCounts:  make([]uint32, windowSize-1),
		windowStart: windowStart,
		WindowSize:  time.Duration(windowSize) * timeUnit,
		Unit:        timeUnit,
	}
}

// Observe adds an event to the window at the current moment in time
func (c *Counter) Observe() {
	c.refreshWindow()
	atomic.AddUint32(&c.crtCount, 1)
}

// Value returns the number of events within the window
func (c *Counter) Value() int {
	c.refreshWindow()

	sum := atomic.LoadUint32(&c.crtCount)
	c.mu.RLock()
	for i := 0; i < len(c.prevCounts); i++ {
		sum += c.prevCounts[i]
	}
	c.mu.RUnlock()

	return int(sum)
}

// refreshWindow ensures the end of the window is on the current time unit
func (c *Counter) refreshWindow() {
	// Truncate current timestamp to match the counter's time unit
	now := time.Now().Truncate(c.Unit)

	c.mu.RLock()
	isCurrentUnitInWindow := now.Sub(c.windowStart) < c.WindowSize
	c.mu.RUnlock()

	if !isCurrentUnitInWindow {
		c.moveWindow(now)
	}
}

// moveWindow moves the window such that its end is on the given time instant
// and removes the counts that fall outside of the window
func (c *Counter) moveWindow(t time.Time) {
	// Round the time instant to the next multiple of time unit such that
	// the window will include this time instant as well
	t = t.Truncate(c.Unit).Add(c.Unit)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Do nothing if the window already covers the given time instant
	if t.Sub(c.windowStart) <= c.WindowSize {
		return
	}

	// Remove the counts that are outside of the current window
	// i.e. remove counts that are older than [t - c.windowSize]
	moveDistance := int((t.Sub(c.windowStart) - c.WindowSize) / c.Unit)
	leftShiftInPlace(c.prevCounts, moveDistance)

	// Move current count into previous counts
	crtCountNewPos := len(c.prevCounts) - moveDistance
	if crtCountNewPos >= 0 {
		c.prevCounts[crtCountNewPos] = atomic.SwapUint32(&c.crtCount, 0)
	} else {
		// Just reset it if it falls outside the window after moving it
		atomic.StoreUint32(&c.crtCount, 0)
	}

	c.windowStart = c.windowStart.Add(time.Duration(moveDistance) * c.Unit)
}

// leftShiftInPlace shifts the elements in s by p positions to the left,
// and inserts zeroes at the right end.
//
// Example:
//   INPUT:  s=[1, 2, 3, 4, 5]; p=2
//   OUTPUT: s=[3, 4, 5, 0, 0]
func leftShiftInPlace(s []uint32, p int) {
	if p <= 0 {
		return
	}

	// Shift elements to the left
	for i := 0; i < len(s)-p; i++ {
		s[i] = s[i+p]
	}

	// "Insert" zeroes at the right end
	start := len(s) - p
	if start < 0 {
		start = 0
	}
	for i := start; i < len(s); i++ {
		s[i] = 0
	}
}
