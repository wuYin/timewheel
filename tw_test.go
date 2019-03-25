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
	tw := NewTimeWheel(100*time.Millisecond, 30)
	start := time.Now() // 时间轮有 1s 钟的误差
	_, doneChs := tw.Repeat(1*time.Second, 4, func() {
		fmt.Println(fmt.Sprintf("spent: %.f", time.Now().Sub(start).Seconds()))
	})
	var wg sync.WaitGroup
	wg.Add(7)
	go func() {
		for _, done := range doneChs {
			if _, ok := <-done; !ok {
				wg.Done()
			}
		}
	}()
	wg.Wait()
}
