package timewheel

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type TimeWheel struct {
	ticker       *time.Ticker
	tickDuration time.Duration     // 每次 tick 时长
	slotNum      int               // 槽的数量，即每轮 tick 总数+1
	curSlot      int               // 当前 slot
	lock         sync.RWMutex      // 保护 idx
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
			tw.lock.Lock()
			prevSlot := tw.slots[t.slot]
			node := prevSlot.tasks.Push(t)
			tw.taskMap[t.idx] = node
			tw.lock.Unlock()
		}
	}
}

// 执行延时任务
func (tw *TimeWheel) After(timeout time.Duration, exec func()) (int64, chan struct{}) {
	t := newTask(timeout, false, exec)
	tw.fillTaskIdx(t)
	tw.taskCh <- t
	return t.idx, t.doneCh
}

// 指定重复任务
func (tw *TimeWheel) Repeat(timeout time.Duration, exec func()) (int64, chan struct{}) {
	t := newTask(timeout, true, exec)
	tw.fillTaskIdx(t)
	tw.taskCh <- t
	return t.idx, t.doneCh
}

// 中途取消延时任务的执行
func (tw *TimeWheel) Cancel(taskId int64) {
	tw.lock.Lock()
	defer tw.lock.Unlock()

	if node, ok := tw.taskMap[taskId]; ok {
		close(node.Value().(*twTask).doneCh) // 直接关闭
		slotIdx := tw.task2Slot(taskId)
		tw.slots[slotIdx].tasks.Remove(node)
		delete(tw.taskMap, taskId)
	}
}

// 填充 task 的 idx
func (tw *TimeWheel) fillTaskIdx(t *twTask) {
	slotIdx := tw.prevSlotIdx()

	tw.lock.Lock()
	taskIdx := tw.slot2Task(slotIdx)
	t.idx = taskIdx
	t.slot = slotIdx
	tw.lock.Unlock()
}

// 执行指定 slot 中的所有任务
func (tw *TimeWheel) handleSlotTasks(idx int) {
	var expNodes []*twNode

	tw.lock.RLock()
	slots := tw.slots[idx]
	for node := slots.tasks.Head(); node != nil; node = node.Next() {
		task := node.Value().(*twTask)
		task.cycles--
		if task.cycles > 0 {
			continue
		}
		// 重复任务重新恢复 cycle
		if task.repeat {
			task.cycles = cycle(task.timeout)
		}

		if !task.repeat {
			expNodes = append(expNodes, node)
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("task exec paic: %v", err) // 出错暂只记录
				}
			}()
			task.exec()               // 任务的执行是异步的
			task.doneCh <- struct{}{} //
			if !task.repeat {
				close(task.doneCh)
			}
		}()
	}
	tw.lock.RUnlock()

	for _, n := range expNodes {
		tw.lock.Lock()
		slots.tasks.Remove(n)                       // 剔除过期任务
		delete(tw.taskMap, n.Value().(*twTask).idx) //
		tw.lock.Unlock()
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

// 获取上一个 slot 的索引
func (tw *TimeWheel) prevSlotIdx() int {
	tw.lock.RLock()         // 注意临界区不要重叠...不然锁会相互等待造成死锁
	defer tw.lock.RUnlock() //

	i := tw.curSlot
	if i > 0 {
		return i - 1
	}
	return tw.slotNum - 1 // 当前已是最后一个 slot
}

func (tw *TimeWheel) String() (s string) {
	for _, slot := range tw.slots {
		if slot.tasks.Size() > 0 {
			s += fmt.Sprintf("[%v]\t", slot.tasks)
		}
	}
	return
}
