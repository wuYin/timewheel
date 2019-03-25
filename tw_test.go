package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestAfter(t *testing.T) {
	tw := NewTimeWheel(1*time.Second, 6)
	start := time.Now()
	_, done := tw.After(2*time.Second, func() {
		fmt.Println(fmt.Sprintf("spent: %d", time.Now().Sub(start).Nanoseconds()/1e6))
	})
	for range done {
	}
}

func TestRepeat(t *testing.T) {
	tw := NewTimeWheel(1*time.Second, 6)
	start := time.Now().Add(1 * time.Second)
	_, done := tw.Repeat(1*time.Second, 2, func() {
		fmt.Println(fmt.Sprintf("spent: %d", time.Now().Sub(start).Nanoseconds()/1e6))
	})
	for range done {
	}
}
