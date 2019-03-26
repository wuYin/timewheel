package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestAfter(t *testing.T) {
	tw := NewTimeWheel(10*time.Millisecond, 6000)
	start := time.Now()
	_, resCh := tw.After(2*time.Second, func() interface{} {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
		return true
	})
	for res := range resCh {
		succ, ok := res.(bool)
		if !ok || succ {
			t.Fail()
		}
	}
}

func TestAfterPoints(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 600)
	points := []int64{1, 2, 4, 8, 16}
	start := time.Now()
	_, resChs := tw.AfterPoints(1*time.Second, points, func() interface{} {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
		return true
	})
	for _, resCh := range resChs {
		for res := range resCh {
			succ, ok := res.(bool)
			if !ok || succ {
				t.Fail()
			}
		}
	}
}

func TestRepeat(t *testing.T) {
	tw := NewTimeWheel(10*time.Millisecond, 6000)
	start := time.Now()
	_, allDoneCh := tw.Repeat(1*time.Second, 5, func() interface{} {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
		return true
	})
	<-allDoneCh
}

func TestCancel(t *testing.T) {
	tw := NewTimeWheel(1*time.Second, 3)
	tid, _ := tw.After(4*time.Second, func() interface{} {
		fmt.Println("after 4s, task executed")
		return true
	})
	time.Sleep(3 * time.Second)
	if !tw.Cancel(tid) {
		t.Fail()
	}
	if len(tw.taskMap) != 0 {
		t.Fail()
	}
}

func TestUpdate(t *testing.T) {
	tw := NewTimeWheel(10*time.Millisecond, 6000)
	start := time.Now()
	tids, _ := tw.Repeat(1*time.Second, 2, func() interface{} {
		fmt.Println(fmt.Sprintf("[origin] spent: %.2fs", time.Now().Sub(start).Seconds()))
		return true
	})
	time.Sleep(2500 * time.Millisecond)
	_, allDoneCh := tw.Update(tids, 1*time.Second, 4, func() interface{} {
		fmt.Println(fmt.Sprintf("[updated] spent: %.2fs", time.Now().Sub(start).Seconds()))
		return true
	})
	<-allDoneCh
}
