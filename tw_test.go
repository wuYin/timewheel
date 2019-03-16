package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeWheel(t *testing.T) {
	tw := NewTimeWheel(1*time.Millisecond, 1000)
	idx1, doneCh1 := tw.After(2*time.Second, func() {
		fmt.Println("after 2 seconds, task1 executed")
	})
	fmt.Println("idx1", idx1)

	idx2, doneCh2 := tw.After(2*time.Second, func() {
		fmt.Println("after 2 seconds, task2 executed")
	})
	fmt.Println("idx2", idx2)

	time.Sleep(10 * time.Millisecond) // 让 task1 有机会执行
	succ := tw.Update(idx1, 3*time.Second, func() {
		fmt.Println("task1 updated, after 3 seconds, task1 executed")
	})
	if !succ {
		t.Fatalf("update task1 failed")
	}

	idx3, doneCh3 := tw.Repeat(3*time.Second, func() {
		fmt.Println("per 3 seconds, task3 executed")
	})
	fmt.Println("idx3", idx3)

	go func() {
		time.Sleep(1 * time.Second) // 延迟 1s 取消 task2
		tw.Cancel(idx2)
	}()

	go func() {
		i := 0
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				i++
				fmt.Printf("----------TICK %d------------\n", i)
			}
		}
	}()

	go func() {
		for {
			fmt.Println(tw)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	if _, ok := <-doneCh1; !ok {
		t.Fatal("task1 has not executed")
	}
	if _, ok := <-doneCh2; ok {
		t.Fatal("task2 has been executed")
	}

	for range doneCh3 {
		// 会一直阻塞在这
	}
}
