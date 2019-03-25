package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestTW(t *testing.T) {
	tw := NewTimeWheel(1*time.Second, 60)
	start := time.Now()
	tw.After(1*time.Second, func() {
		fmt.Println(time.Now().Sub(start).Nanoseconds() / 1e6)
	})
	select {}
}
