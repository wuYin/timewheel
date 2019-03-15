package timewheel

import (
	"container/list"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type TimeWheel struct {
	ticker       *time.Ticker
	tickDuration time.Duration           // 每次 tick 时长
	slotNum      int                     // 槽的数量，即每轮 tick 总数+1
	curSlot      int                     // 当前 slot
	lock         sync.RWMutex            // 保护 idx
	slots        []*twSlot               // 全部槽
	taskMap      map[int64]*list.Element // taskId -> task Addr
	incrId       int64                   // 自增 id
}

func NewTimeWheel(tickDuration time.Duration, slotNum int) *TimeWheel {
	tw := &TimeWheel{
		ticker:       time.NewTicker(tickDuration),
		tickDuration: tickDuration,
		slotNum:      slotNum,
		slots:        make([]*twSlot, 0, slotNum),
		taskMap:      make(map[int64]*list.Element),
	}
	for i := 0; i < slotNum; i++ {
		tw.slots = append(tw.slots, newSlot(i))
	}

	go func() {
		i := 0 // slot 编号
		for {
			i %= tw.slotNum
			<-tw.ticker.C // ticking...

			tw.lock.Lock()
			tw.curSlot = i
			tw.lock.Unlock()

			tw.handleSlotTasks(i) // 不会阻塞
			i++
		}
	}()

	return tw
}

// 执行延时任务
func (tw *TimeWheel) After(timeout time.Duration, exec func()) (int64, chan struct{}) {
	prevIdx := tw.prevSlotIdx()
	cycles := int(int64(timeout) / (int64(tw.tickDuration) * int64(tw.slotNum)))

	tw.lock.Lock()
	defer tw.lock.Unlock()

	idx := tw.slot2Task(prevIdx)
	task := newTask(idx, cycles, exec)

	prevSlot := tw.slots[prevIdx]
	node := prevSlot.tasks.PushFront(task)

	tw.taskMap[idx] = node
	return idx, task.doneCh
}

// 中途取消延时任务的执行
func (tw *TimeWheel) Cancel(taskId int64) {
	tw.lock.Lock()
	defer tw.lock.Unlock()

	if node, ok := tw.taskMap[taskId]; ok {
		close(node.Value.(*twTask).doneCh) // 直接关闭
		slotIdx := tw.task2Slot(taskId)
		tw.slots[slotIdx].tasks.Remove(node)
		delete(tw.taskMap, taskId)
	}
}

// 执行指定 slot 中的所有任务
func (tw *TimeWheel) handleSlotTasks(idx int) {
	var expNodes []*list.Element

	tw.lock.RLock()
	slots := tw.slots[idx]
	for node := slots.tasks.Back(); node != nil; node = node.Prev() {
		task := node.Value.(*twTask)
		task.cycles--
		if task.cycles > 0 {
			continue
		}

		expNodes = append(expNodes, node)
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("task exec paic: %v", err) // 出错暂只记录
				}
			}()
			task.exec()               // 任务的执行是异步的
			task.doneCh <- struct{}{} //
			close(task.doneCh)
		}()
	}
	tw.lock.RUnlock()

	for _, n := range expNodes {
		tw.lock.Lock()
		slots.tasks.Remove(n)                     // 剔除过期任务
		delete(tw.taskMap, n.Value.(*twTask).idx) //
		tw.lock.Unlock()
	}
}

// 在指定 slot 中无重复生成新 task id
func (tw *TimeWheel) slot2Task(slotIdx int) int64 {
	atomic.AddInt64(&tw.incrId, 1)
	return int64(slotIdx)<<32 | tw.incrId>>32 // 使用 incrId 将 taskIdx 进行 shuffle 且保证去重
}

// 反向获取 task 所在的 slot
func (tw *TimeWheel) task2Slot(taskIdx int64) int {
	return int(taskIdx >> 32)
}

// 获取上一个 slot 的索引
func (tw *TimeWheel) prevSlotIdx() int {
	tw.lock.RLock()
	defer tw.lock.RUnlock()

	i := tw.curSlot
	if i > 0 {
		return i - 1
	}
	return tw.slotNum - 1 // 当前已是最后一个 slot
}
