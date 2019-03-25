package timewheel

import (
	"fmt"
	"time"
)

// 每个 slot 链表中的 task
type twTask struct {
	id       int64         // 在 slot 中的索引位置
	slotIdx  int           // 所属 slot
	interval time.Duration // 任务执行间隔
	cycles   int64         // 延迟指定圈后执行
	do       func()        // 执行任务
	doneCh   chan struct{} // 通知任务执行结束
	repeat   int64         // 任务重复执行次数
}

func newTask(interval time.Duration, repeat int64, do func()) *twTask {
	return &twTask{
		interval: interval,
		cycles:   cycle(interval),
		repeat:   repeat,
		do:       do,
		doneCh:   make(chan struct{}, 1), // buffed
	}
}

// 计算 timeout 应在第几圈被执行
func cycle(interval time.Duration) (n int64) {
	n = 1 + int64(interval)/cycleCost
	return
}

func (t *twTask) String() string {
	return fmt.Sprintf("[slot]:%d [interval]:%.fs [repeat]:%d [cycle]:%dth [idx]:%d ",
		t.slotIdx, t.interval.Seconds(), t.repeat, t.cycles, t.id)
}
