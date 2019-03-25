package timewheel

import (
	"fmt"
	"sync"
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
	tw := NewTimeWheel(10*time.Millisecond, 3)
	start := time.Now()
	_, doneChs := tw.Repeat(1*time.Second, 4, func() {
		fmt.Println(fmt.Sprintf("spent: %.fs", time.Now().Sub(start).Seconds()))
	})
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		for _, done := range doneChs {
			if _, ok := <-done; !ok {
				wg.Done()
			}
		}
	}()
	wg.Wait()
}
