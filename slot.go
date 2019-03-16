package timewheel

// 时间槽
type twSlot struct {
	idx   int
	tasks *twList
}

func newSlot(idx int) *twSlot {
	return &twSlot{idx: idx, tasks: newList()}
}
