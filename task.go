package timewheel

import (
	"fmt"
	"time"
)

// 每个 slot 链表中的 task
type twTask struct {
	slot    int           // 所属 slot
	idx     int64         // 在 slot 中的索引位置
	timeout time.Duration // 任务执行间隔
	cycles  int           // 延迟指定圈后执行
	exec    func()        // 执行任务
	doneCh  chan struct{} // 通知任务执行结束
	repeat  bool          // 任务是否需要重复执行
}

func newTask(timeout time.Duration, repeat bool, exec func()) *twTask {
	return &twTask{
		timeout: timeout,
		cycles:  cycle(timeout),
		repeat:  repeat,
		exec:    exec,
		doneCh:  make(chan struct{}, 1), // buffed
	}
}

// 计算 timeout 应在第几圈被执行
func cycle(timeout time.Duration) (n int) {
	n = int(timeout / cycleCost)
	return
}

func (t *twTask) String() string {
	idx := fmt.Sprintf("%d", t.idx)
	return fmt.Sprintf("[slot]:%d [time]:%.f [idx]:%s [cycle]:%d", t.slot, t.timeout.Seconds(), idx[len(idx)-2:], t.cycles)
}
