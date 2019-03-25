package timewheel

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

type TimeWheel struct {
	ticker       *time.Ticker
	tickDuration time.Duration     // 每次 tick 时长
	slotNum      int               // 槽的数量，即每轮 tick 总数+1
	curSlot      int               // 当前 slot
	slots        []*twSlot         // 全部槽
	taskMap      map[int64]*twNode // taskId -> task Addr
	incrId       int64             // 自增 id
	taskCh       chan *twTask      // 传递 task
}

var (
	cycleCost time.Duration // 周期耗时
)

func NewTimeWheel(tickDuration time.Duration, slotNum int) *TimeWheel {
	tw := &TimeWheel{
		ticker:       time.NewTicker(tickDuration),
		tickDuration: tickDuration,
		slotNum:      slotNum,
		slots:        make([]*twSlot, 0, slotNum),
		taskMap:      make(map[int64]*twNode),
		taskCh:       make(chan *twTask, 100),
	}
	cycleCost = tw.tickDuration * time.Duration(tw.slotNum)
	for i := 0; i < slotNum; i++ {
		tw.slots = append(tw.slots, newSlot(i))
	}

	go tw.run()
	return tw
}

// 接收并运行定时任务
func (tw *TimeWheel) run() {
	idx := 0
	for {
		select {
		case <-tw.ticker.C:
			idx %= tw.slotNum
			tw.curSlot = idx
			tw.handleSlotTasks(idx)
			idx++
		case t := <-tw.taskCh:
			fmt.Println(t)
			slot := tw.slots[t.slotIdx]
			node := slot.tasks.Push(t)
			tw.taskMap[t.id] = node
		}
	}
}

// 执行延时任务
func (tw *TimeWheel) After(timeout time.Duration, do func()) (int64, chan struct{}) {
	t := newTask(timeout, 1, do)
	tw.locate(t, t.interval)
	tw.taskCh <- t
	return t.id, t.doneCh
}

// 指定重复任务
func (tw *TimeWheel) Repeat(interval time.Duration, repeatN int64, do func()) (int64, chan struct{}) {
	intervalSum := repeatN * int64(interval)
	trip := intervalSum / int64(cycleCost) // 往返多少趟

	if trip > 0 {
		lap := interval
		for cur := time.Duration(0); cur < cycleCost; cur += interval { // 每隔 interval 放置执行 trip 次的 task
			t := newTask(interval, trip, do)
			tw.locate(t, lap)
			tw.taskCh <- t
			lap += interval
		}
	}

	var lastId int64
	var lastDone chan struct{}
	lap := interval
	remain := (intervalSum % int64(cycleCost)) / int64(tw.tickDuration)
	for i := 0; i < int(remain); i++ {
		t := newTask(interval, 1, do)
		tw.locate(t, lap)
		tw.taskCh <- t
		lastId, lastDone = t.id, t.doneCh
		lap += interval
	}

	return lastId, lastDone
}

// 找准 task 在时间轮中的位置
func (tw *TimeWheel) locate(t *twTask, interval time.Duration) {
	t.slotIdx = tw.convSlotIdx(interval)
	t.id = tw.slot2Task(t.slotIdx)
}

// 执行指定 slot 中的所有任务
func (tw *TimeWheel) handleSlotTasks(idx int) {
	var expNodes []*twNode

	slot := tw.slots[idx]
	for node := slot.tasks.Head(); node != nil; node = node.Next() {
		task := node.Value().(*twTask)
		task.cycles--
		if task.cycles > 0 {
			continue
		}
		// 重复任务重新恢复 cycle
		if task.repeat > 0 {
			task.cycles = cycle(task.interval)
			task.repeat--
		}

		// 不重复任务或重复任务最后一次执行都将移除
		if task.repeat == 0 {
			expNodes = append(expNodes, node)
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("task exec paic: %v", err) // 出错暂只记录
				}
			}()
			task.do()                 // 任务的执行是异步的
			task.doneCh <- struct{}{} // 通知执行完毕
			if task.repeat == 0 {
				close(task.doneCh)
			}
		}()
	}

	for _, n := range expNodes {
		slot.tasks.Remove(n)                       // 剔除过期任务
		delete(tw.taskMap, n.Value().(*twTask).id) //
	}
}

// 在指定 slot 中无重复生成新 task id
func (tw *TimeWheel) slot2Task(slotIdx int) int64 {
	return int64(slotIdx)<<32 + atomic.AddInt64(&tw.incrId, 1) // 保证去重优先
}

// 反向获取 task 所在的 slot
func (tw *TimeWheel) task2Slot(taskIdx int64) int {
	return int(taskIdx >> 32)
}

// 将 task 的 interval 计算到指定的 slot 中
func (tw *TimeWheel) convSlotIdx(interval time.Duration) int {
	timeGap := interval % cycleCost
	slotGap := int(timeGap / tw.tickDuration)
	return tw.curSlot + slotGap%tw.slotNum // 误差来源
}

func (tw *TimeWheel) String() (s string) {
	for _, slot := range tw.slots {
		if slot.tasks.Size() > 0 {
			s += fmt.Sprintf("[%v]\t", slot.tasks)
		}
	}
	return
}
