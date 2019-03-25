package timewheel

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type TimeWheel struct {
	ticker  *time.Ticker
	tickGap time.Duration     // 每次 tick 时长
	slotNum int               // slot 数量
	curSlot int               // 当前 slot 序号
	slots   []*twSlot         // 槽数组
	taskMap map[int64]*twNode // taskId -> taskPtr
	incrId  int64             // 自增 id
	taskCh  chan *twTask      // task 缓冲 channel
	lock    sync.RWMutex      // 数据读写锁
}

var cycleCost int64 // 周期耗时

// 生成 slotNum 个以 tickGap 为时间间隔的时间轮
func NewTimeWheel(tickGap time.Duration, slotNum int) *TimeWheel {
	tw := &TimeWheel{
		ticker:  time.NewTicker(tickGap),
		tickGap: tickGap,
		slotNum: slotNum,
		slots:   make([]*twSlot, 0, slotNum),
		taskMap: make(map[int64]*twNode),
		taskCh:  make(chan *twTask, 100),
		lock:    sync.RWMutex{},
	}
	cycleCost = int64(tw.tickGap * time.Duration(tw.slotNum))
	for i := 0; i < slotNum; i++ {
		tw.slots = append(tw.slots, newSlot(i))
	}

	go tw.turn()

	return tw
}

// 执行延时任务
func (tw *TimeWheel) After(timeout time.Duration, do func()) (int64, chan struct{}) {
	if timeout <= 0 {
		return -1, nil
	}

	t := newTask(timeout, 1, do)
	tw.locate(t, t.interval, false)
	tw.taskCh <- t
	return t.id, t.doneCh
}

// 执行重复任务
func (tw *TimeWheel) Repeat(interval time.Duration, repeatN int64, do func()) ([]int64, chan struct{}) {
	if interval <= 0 || repeatN < 1 {
		return nil, nil
	}

	costSum := repeatN * int64(interval) // 全部任务耗时
	cycleSum := costSum / cycleCost      // 全部任务执行总圈数
	trip := cycleSum / cycle(interval)   // 每个任务多少圈才执行一次

	var tids []int64
	var doneChs []chan struct{}
	if trip > 0 {
		gap := interval
		for step := int64(0); step < cycleCost; step += int64(interval) { // 每隔 interval 放置执行 trip 次的 task
			t := newTask(interval, trip, do)
			tw.locate(t, gap, false)
			tw.taskCh <- t
			gap += interval
			tids = append(tids, t.id)
			doneChs = append(doneChs, t.doneCh)
		}
	}

	// 计算余下几个任务时需重头开始计算
	gap := time.Duration(0)
	remain := (costSum % cycleCost) / int64(interval)
	for i := 0; i < int(remain); i++ {
		t := newTask(interval, 1, do)
		t.cycles = trip + 1
		tw.locate(t, gap, true)
		tw.taskCh <- t
		gap += interval
		tids = append(tids, t.id)
		doneChs = append(doneChs, t.doneCh)
	}

	allDone := make(chan struct{}, 1)
	go func(doneChs []chan struct{}) {
		for _, ch := range doneChs {
			for range ch {
			}
		}
		allDone <- struct{}{} // 等待全部子任务完成
	}(doneChs)
	return tids, allDone
}

// 更新任务
func (tw *TimeWheel) Update(tids []int64, interval time.Duration, repeatN int64, do func()) ([]int64, chan struct{}) {
	if len(tids) == 0 || interval <= 0 || repeatN < 1 {
		return nil, nil
	}

	if repeatN == 1 {
		if !tw.Cancel(tids[0]) {
			// return nil, nil // 按需处理
		}
		newTid, doneCh := tw.After(interval, do)
		return []int64{newTid}, doneCh
	}

	// 重复任务需逐个全部取消
	for _, tid := range tids {
		if !tw.Cancel(tid) {
			// return nil, nil // 按需处理
		}
	}
	return tw.Repeat(interval, repeatN, do)
}

// 取消任务
func (tw *TimeWheel) Cancel(tid int64) bool {
	tw.lock.Lock()
	defer tw.lock.Unlock()

	node, ok := tw.taskMap[tid]
	if !ok {
		return false // 任务已执行完毕或不存在
	}

	t := node.value.(*twTask)
	t.doneCh <- struct{}{}
	close(t.doneCh) // 避免资源泄漏

	slot := tw.slots[t.slotIdx]
	slot.tasks.remove(node)
	delete(tw.taskMap, tid)
	return true
}

// 接收 task 并定时运行 slot 中的任务
func (tw *TimeWheel) turn() {
	idx := 0
	for {
		select {
		case <-tw.ticker.C:
			idx %= tw.slotNum
			tw.lock.Lock()
			tw.curSlot = idx // 锁粒度要细，不要重叠
			tw.lock.Unlock()
			tw.handleSlotTasks(idx)
			idx++
		case t := <-tw.taskCh:
			tw.lock.Lock()
			// fmt.Println(t)
			slot := tw.slots[t.slotIdx]
			tw.taskMap[t.id] = slot.tasks.push(t)
			tw.lock.Unlock()
		}
	}
}

// 计算 task 所在 slot 的编号
func (tw *TimeWheel) locate(t *twTask, gap time.Duration, restart bool) {
	tw.lock.Lock()
	defer tw.lock.Unlock()
	if restart {
		t.slotIdx = tw.convSlotIdx(gap)
	} else {
		t.slotIdx = tw.curSlot + tw.convSlotIdx(gap)
	}
	t.id = tw.slot2Task(t.slotIdx)
}

// 执行指定 slot 中的所有任务
func (tw *TimeWheel) handleSlotTasks(idx int) {
	var expNodes []*twNode

	tw.lock.RLock()
	slot := tw.slots[idx]
	for node := slot.tasks.head; node != nil; node = node.next {
		task := node.value.(*twTask)
		task.cycles--
		if task.cycles > 0 {
			continue
		}
		// 重复任务恢复 cycle
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
			if task.do != nil {
				task.do()
			}
			task.doneCh <- struct{}{} // 通知执行完毕
			if task.repeat == 0 {
				close(task.doneCh)
			}
		}()
	}
	tw.lock.RUnlock()

	tw.lock.Lock()
	for _, n := range expNodes {
		slot.tasks.remove(n)                     // 剔除过期任务
		delete(tw.taskMap, n.value.(*twTask).id) //
	}
	tw.lock.Unlock()
}

// 在指定 slot 中无重复生成新 task id
func (tw *TimeWheel) slot2Task(slotIdx int) int64 {
	return int64(slotIdx)<<32 + atomic.AddInt64(&tw.incrId, 1) // 保证去重优先
}

// 反向获取 task 所在的 slot
func (tw *TimeWheel) task2Slot(taskIdx int64) int {
	return int(taskIdx >> 32)
}

// 将指定间隔计算到指定的 slot 中
func (tw *TimeWheel) convSlotIdx(gap time.Duration) int {
	timeGap := gap % time.Duration(cycleCost)
	slotGap := int(timeGap / tw.tickGap)
	return int(slotGap % tw.slotNum)
}

func (tw *TimeWheel) String() (s string) {
	for _, slot := range tw.slots {
		if slot.tasks.size > 0 {
			s += fmt.Sprintf("[%v]\t", slot.tasks)
		}
	}
	return
}
