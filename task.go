package timewheel

import (
	"fmt"
	"time"
)

type doTask func() interface{}

// 每个 slot 链表中的 task
type twTask struct {
	id       int64            // 在 slot 中的索引位置
	slotIdx  int              // 所属 slot
	interval time.Duration    // 任务执行间隔
	cycles   int64            // 延迟指定圈后执行
	do       doTask           // 执行任务
	resCh    chan interface{} // 传递任务执行结果
	repeat   int64            // 任务重复执行次数
}

func newTask(interval time.Duration, repeat int64, do func() interface{}) *twTask {
	return &twTask{
		interval: interval,
		cycles:   cycle(interval),
		repeat:   repeat,
		do:       do,
		resCh:    make(chan interface{}, 1),
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
