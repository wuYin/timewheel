package timewheel

// 每个 slot 链表中的 task
type twTask struct {
	idx    int64         // 在 slot 中的索引位置
	cycles int           // 延迟指定圈后执行
	exec   func()        // 执行任务
	doneCh chan struct{} // 通知任务执行结束
}

func newTask(idx int64, cycles int, exec func()) *twTask {
	return &twTask{
		idx:    idx,
		exec:   exec,
		cycles: cycles,
		doneCh: make(chan struct{}, 1), // buffed
	}
}
