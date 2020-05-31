package hops

import (
	"reflect"
	"testing"
	"time"
)

func TestMoveWindow(t *testing.T) {
	var newCounter = func() *Counter {
		c := NewCounter(5, time.Second)
		c.prevCounts = []uint32{1, 2, 3, 4}
		c.crtCount = 99
		return c
	}

	tests := map[string]struct {
		timeUnitsFromWindowEnd int
		expectedPrevCounts     []uint32
	}{
		"one_unit": {
			1,
			[]uint32{2, 3, 4, 99},
		},
		"two_units": {
			2,
			[]uint32{3, 4, 99, 0},
		},
		"keep_only_current_unit": {
			4,
			[]uint32{99, 0, 0, 0},
		},
		"just_outside_of_the_window": {
			5,
			[]uint32{0, 0, 0, 0},
		},
		"way_outside_of_the_window": {
			10,
			[]uint32{0, 0, 0, 0},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := newCounter()
			windowEnd := c.windowStart.Add(c.WindowSize - c.Unit)

			// Simulate a couple of time units have passed since the counter was last used
			unitsPassed := time.Duration(tt.timeUnitsFromWindowEnd) * c.Unit
			c.moveWindow(windowEnd.Add(unitsPassed))

			if !reflect.DeepEqual(c.prevCounts, tt.expectedPrevCounts) {
				t.Errorf("Old counts were not removed: expected: %v, got: %v",
					tt.expectedPrevCounts, c.prevCounts)
			}
			if c.crtCount != 0 {
				t.Errorf("Current count was not reset. Got: %d", c.crtCount)
			}
		})
	}
}

func TestLeftShiftInPlace(t *testing.T) {
	tests := map[string]struct {
		shift int
		slice []uint32
		want  []uint32
	}{
		"shift_one": {
			1,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{2, 3, 4, 5, 0},
		},
		"shift_two": {
			2,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{3, 4, 5, 0, 0},
		},
		"all_elements_out": {
			10,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{0, 0, 0, 0, 0},
		},
		"shift_by_slice_length": {
			5,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{0, 0, 0, 0, 0},
		},
		"keep_the_rightmost_element": {
			4,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{5, 0, 0, 0, 0},
		},
		"one_element_slice": {
			1,
			[]uint32{1},
			[]uint32{0},
		},
		"empty_slice": {
			1,
			[]uint32{},
			[]uint32{},
		},
		"no_shift": {
			0,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{1, 2, 3, 4, 5},
		},
		"negative_shift": {
			-3,
			[]uint32{1, 2, 3, 4, 5},
			[]uint32{1, 2, 3, 4, 5},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			leftShiftInPlace(tt.slice, tt.shift)
			if !reflect.DeepEqual(tt.slice, tt.want) {
				t.Errorf("expected: %v, got: %v", tt.want, tt.slice)
			}
		})
	}
}
