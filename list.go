package timewheel

import "fmt"

type twNode struct {
	value        interface{}
	prev, next *twNode
}

func newNode(v interface{}) *twNode {
	return &twNode{value: v}
}

type twList struct {
	head, tail *twNode
	size       int
}

func newList() *twList {
	return new(twList)
}

func (l *twList) push(v interface{}) *twNode {
	n := newNode(v)
	if l.head == nil {
		l.head, l.tail = n, n
		l.size++
		return n
	}

	n.prev = l.tail
	n.next = nil

	l.tail.next = n
	l.tail = n
	l.size++
	return n
}

func (l *twList) remove(n *twNode) {
	if n == nil {
		return
	}

	prev, next := n.prev, n.next
	if prev == nil {
		l.head = next
	} else {
		prev.next = next
	}

	if next == nil {
		l.tail = prev
	} else {
		next.prev = prev
	}
	n = nil // 主动释放内存
	l.size--
}

func (l *twList) String() (s string) {
	s = fmt.Sprintf("[%d]: ", l.size)
	for cur := l.head; cur != nil; cur = cur.next {
		s += fmt.Sprintf("%v <-> ", cur.value)
	}
	s += "<nil>"

	return s
}
