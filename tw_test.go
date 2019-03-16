package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeWheel(t *testing.T) {
	tw := NewTimeWheel(1*time.Millisecond, 1000)
	idx1, waitCh1 := tw.After(2*time.Second, func() {
		fmt.Println("cycle 1th and waited 2 seconds...")
	})
	fmt.Println("idx1", idx1)

	idx2, waitCh2 := tw.After(2*time.Second, func() {
		fmt.Println("cycle 1th and waited 2 seconds...")
	})
	fmt.Println("idx2", idx2)

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
				fmt.Println("tick:", i)
			}
		}
	}()

	if _, ok := <-waitCh1; !ok {
		t.Fatal("task1 has not executed")
	}
	if _, ok := <-waitCh2; ok {
		t.Fatal("task2 has been executed")
	}
}
