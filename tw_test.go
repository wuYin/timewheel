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

// 存在误差 10ms
func TestRepeat(t *testing.T) {
	tw := NewTimeWheel(10*time.Millisecond, 300)
	start := time.Now()
	_, doneChs := tw.Repeat(1*time.Second, 5, func() {
		fmt.Println(fmt.Sprintf("spent: %.fs", time.Now().Sub(start).Seconds()))
	})
	for _, done := range doneChs {
		for range done {
		}
	}
}
