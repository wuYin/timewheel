package timewheel

// 时间槽
type twSlot struct {
	id    int
	tasks *twList
}

func newSlot(id int) *twSlot {
	return &twSlot{id: id, tasks: newList()}
}
