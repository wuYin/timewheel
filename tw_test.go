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
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
	})
	for range done {
	}
}

func TestAfterPoints(t *testing.T) {
	tw := NewTimeWheel(100*time.Millisecond, 600)
	points := []int64{0, 2, 4, 8, 16}
	start := time.Now()
	_, allDone := tw.AfterPoints(1*time.Second, points, func() {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
	})
	<-allDone
}

func TestRepeat(t *testing.T) {
	tw := NewTimeWheel(1*time.Second, 3)
	start := time.Now()
	_, allDoneCh := tw.Repeat(1*time.Second, 5, func() {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
	})
	<-allDoneCh
}

func TestCancel(t *testing.T) {
	tw := NewTimeWheel(1*time.Second, 3)
	tid, _ := tw.After(4*time.Second, func() {
		fmt.Println("after 4s, task executed")
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
	tw := NewTimeWheel(1*time.Second, 3)
	start := time.Now()
	tids, _ := tw.Repeat(1*time.Second, 4, func() {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
	})
	time.Sleep(2500 * time.Millisecond)
	_, allDoneCh := tw.Update(tids, 1*time.Second, 2, func() {
		fmt.Println(fmt.Sprintf("spent: %.2fs", time.Now().Sub(start).Seconds()))
	})
	<-allDoneCh
}
