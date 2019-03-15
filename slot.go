package timewheel

import "container/list"

// 时间槽
type twSlot struct {
	idx   int
	tasks *list.List
}

func newSlot(idx int) *twSlot {
	return &twSlot{idx: idx, tasks: list.New()}
}
