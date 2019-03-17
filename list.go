package timewheel

import "fmt"

type twNode struct {
	val        interface{}
	prev, next *twNode
}

func newNode(v interface{}) *twNode {
	return &twNode{val: v}
}

func (n *twNode) Value() interface{} {
	return n.val
}

func (n *twNode) Prev() *twNode {
	return n.prev
}

func (n *twNode) Next() *twNode {
	return n.next
}

type twList struct {
	head, tail *twNode
	size       int
}

func newList() *twList {
	return new(twList)
}

func (l *twList) Head() *twNode {
	return l.head
}

func (l *twList) Tail() *twNode {
	return l.tail
}

func (l *twList) Size() int {
	return l.size
}

func (l *twList) Push(v interface{}) *twNode {
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

func (l *twList) Remove(n *twNode) {
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
	s = fmt.Sprintf("[%d]: ", l.Size())
	for cur := l.Head(); cur != nil; cur = cur.Next() {
		s += fmt.Sprintf("%v <-> ", cur.Value())
	}
	s += "<nil>"

	return s
}
